package rpcserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Info struct {
	Pid int `json:"pid"`
}

func (s *RPCServer) rpcInfo(w http.ResponseWriter, r *http.Request) {
	info := Info{
		Pid: os.Getpid(),
	}

	if err := json.NewEncoder(w).Encode(info); err != nil {
		s.log.Error(err)
		http.Error(w, fmt.Sprintf("internal server error: %s", err.Error()), http.StatusInternalServerError)
	}
}
