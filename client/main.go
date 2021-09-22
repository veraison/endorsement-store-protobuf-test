package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	pb "github.com/veraison/protobuf-test/endorsementapi"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewEndorsementFetcherClient(conn)

	evidence, err := structpb.NewStruct(map[string]interface{}{
		"sw_components": []interface{}{
			map[string]interface{}{

				"impl_id":     "5051525354555657505152535455565750515253545556575051525354555657",
				"prod_id":     "acme.example/rr-trap",
				"inst_id":     "01a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3",
				"type":        "BL",
				"signer_id":   "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
				"version":     "3.4.2",
				"description": "TF-M_SHA256MemPreXIP",
				"measurement": "76543210fedcba9817161514131211101f1e1d1c1b1a1916",
			},
			map[string]interface{}{
				"impl_id":     "5051525354555657505152535455565750515253545556575051525354555657",
				"prod_id":     "acme.example/rr-trap",
				"inst_id":     "01a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3",
				"type":        "M1",
				"signer_id":   "76543210fedcba9817161514131211101f1e1d1c1b1a1918",
				"version":     "1.2",
				"measurement": "76543210fedcba9817161514131211101f1e1d1c1b1a1917",
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	args := &pb.EndorsementArgs{
		Id: &pb.EndorsementID{
			Type: "psa",
			Parts: map[string]string{
				"impl_id": "5051525354555657505152535455565750515253545556575051525354555657",
				"inst_id": "01a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3a0a1a2a3",
			},
		},
		Evidence: &pb.Evidence{Value: evidence},
	}

	reply, err := client.GetEndorsements(context.Background(), args)
	if err != nil {
		log.Fatalf("query error: %v", err)
	}

	log.Printf("Got reply: %v", reply)
}
