To build and run:

    make
    ./build/server ./test/endorsement.db

This will start the GRPC server listening on port 50051. Then, in a different
terminal:

    ./build/client
