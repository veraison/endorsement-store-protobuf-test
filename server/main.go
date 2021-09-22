package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	structpb "google.golang.org/protobuf/types/known/structpb"

	pb "github.com/veraison/protobuf-test/endorsementapi"
)

type fetcherServer struct {
	pb.UnimplementedEndorsementFetcherServer

	db   *sql.DB
	path string
}

func (s *fetcherServer) Init(path string) error {
	dbConfig := fmt.Sprintf("file:%s?cache=shared", path)
	db, err := sql.Open("sqlite3", dbConfig)
	if err != nil {
		return err
	}

	s.db = db
	s.path = path

	return nil
}

func (s *fetcherServer) Close() error {
	return s.db.Close()
}

type EndorsementID pb.EndorsementID

func (e *EndorsementID) Get(name string) (interface{}, error) {
	value, ok := e.Parts[name]
	if !ok {
		return nil, fmt.Errorf("Could not find property %q", name)
	}

	return value, nil
}

func (e *EndorsementID) GetString(name string) (string, error) {
	value, err := e.Get(name)
	if err != nil {
		return "", err
	}
	switch t := value.(type) {
	case string:
		return t, nil
	default:
		return "", fmt.Errorf("Value for %q is of type %T, not string", name, t)
	}
}

func (s fetcherServer) GetEndorsements(
	ctx context.Context,
	args *pb.EndorsementArgs,
) (*pb.EndorsementReply, error) {
	data, err := structpb.NewStruct(map[string]interface{}{
		"type":        "M3",
		"kid":         "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
		"version":     "1",
		"measurement": "76543210fedcba9817161514131211101f1e1d1c1b1a1919",
	})

	if err != nil {
		log.Fatalf("could not create endorsements struct: %v", err)
	}

	endID := EndorsementID(*args.Id)

	implID, err := endID.GetString("impl_id")
	if err != nil {
		return nil, fmt.Errorf("could not get impl_id")
	}

	instID, err := endID.GetString("inst_id")
	if err != nil {
		return nil, fmt.Errorf("could not get inst_id")
	}

	// convert the slice of SW component entries into an associative array, mapping
	// the mesurement to the associated metadata.
	evSwComps := make(map[string]map[string]interface{})
	// TODO: need more type chacking here....

	rawComps := args.Evidence.Value.AsMap()["sw_components"]
	for _, comp := range rawComps.([]interface{}) {
		c := comp.(map[string]interface{})
		evSwComps[c["measurement"].(string)] = c
	}

	rows, err := s.db.Query(
		"select measurement, type, version, signer_id "+
			"from psa_endorsements "+
			"where impl_id = ? and inst_id = ?",
		implID,
		instID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var measurement string
	var type_ string
	var version string
	var signerID string

	var tv pb.TrustVector
	tv.HardwareAuthenticity = pb.Status_SUCCESS
	tv.CertificationStatus = pb.Status_UNKNOWN
	tv.ConfigIntegrity = pb.Status_UNKNOWN
	tv.RuntimeIntegrity = pb.Status_UNKNOWN
	tv.SoftwareUpToDateness = pb.Status_UNKNOWN

	tv.SoftwareIntegrity = pb.Status_SUCCESS
	for rows.Next() {
		err := rows.Scan(&measurement, &type_, &version, &signerID)
		if err != nil {
			return nil, err
		}

		evidence, ok := evSwComps[measurement]
		if !ok {
			tv.SoftwareIntegrity = pb.Status_FAILURE
			break
		}

		if (type_ != evidence["type"]) || (signerID != evidence["signer_id"]) ||
			(version != evidence["version"]) {
			tv.SoftwareIntegrity = pb.Status_FAILURE
			break
		}
	}

	resp := pb.EndorsementReply{
		TrustVector:  &tv,
		Endorsements: data,
	}
	log.Printf("sending: %v", &resp)
	return &resp, nil

}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: server PATH_TO_SQLITE_DB")
	}
	dbpath := os.Args[1]

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("could not create listener: %v", err)
	}

	fetcher := fetcherServer{}
	if err := fetcher.Init(dbpath); err != nil {
		log.Fatalf("could not init store: %v", err)
	}
	defer fetcher.Close()

	grpcServer := grpc.NewServer()
	pb.RegisterEndorsementFetcherServer(grpcServer, &fetcher)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("server error: %v", err)
	}

}
