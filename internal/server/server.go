package server

import (
	"fmt"
	"net/http"

	"github.com/ShinoYasx/Slurmer/internal/appconfig"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
)

type Server struct {
	Config      *appconfig.Config
	slurmClient *slurm.Client
}

func (srv *Server) Listen() error {
	slurmClient, err := slurm.NewClient(srv.Config.Slurmrest.URI)
	if err != nil {
		return err
	}
	srv.slurmClient = slurmClient

	addr := fmt.Sprintf("%s:%d", srv.Config.Slurmer.IP, srv.Config.Slurmer.Port)
	fmt.Printf("Server listening on %s\n", addr)
	http.ListenAndServe(addr, nil)
	return nil
}
