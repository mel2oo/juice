package main

import (
	"context"
	"fmt"
	"time"

	pb "github.com/switch-li/juice/examples/greeter"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial(":8890", grpc.WithInsecure())
	if err != nil {
		fmt.Println("grpc server not alive", err)
		return
	}

	defer conn.Close()

	c := pb.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "jack"})
	if err != nil {
		fmt.Println("say hello error", err)
		return
	}

	fmt.Println(r.GetMessage())

	return
}
