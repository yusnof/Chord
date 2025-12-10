package main

import (
	"math/big"
)

type PingRequest struct {
	Message string
}

type PingResponse struct {
	Message string
}

type NodeInformationRequest struct {
	ID   *big.Int
	IP   string
	Port int
}

type NodeInformationResponse struct {
	ID   *big.Int
	IP   string
	Port int
}

type GetFileRequest struct {
	Filename string
}

type GetFileResponse struct {
	Found   bool
	Content string 
}

type StoreFileRequest struct {
	Filename string
	Content  []byte // raw file bytes
}

type StoreFileResponse struct {
	Success bool
	Message string
}
