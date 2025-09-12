package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/c0rydoras/folien/internal/server"
	"github.com/spf13/cobra"
)

var (
	host    string
	port    int
	keyPath string
)

// serveCmd is the command for serving the presentation. It starts the slides
// server allowing for connections.
var serveCmd = &cobra.Command{
	Use:     "serve <file.md>",
	Aliases: []string{"server"},
	Short:   "Start an SSH server to run folien",
	Args:    cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Help()
		}

		k := os.Getenv("FOLIEN_SERVER_KEY_PATH")
		if k != "" {
			keyPath = k
		}
		h := os.Getenv("FOLIEN_SERVER_HOST")
		if h != "" {
			host = h
		}
		p := os.Getenv("FOLIEN_SERVER_PORT")
		if p != "" {
			port, _ = strconv.Atoi(p)
		}

		presentation, err := newModel(args[0])
		if err != nil {
			return err
		}

		s, err := server.NewServer(keyPath, host, port, presentation)
		if err != nil {
			return err
		}

		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		log.Printf("Starting Folien server on %s:%d", host, port)
		go func() {
			if err = s.Start(); err != nil {
				log.Fatalln(err)
			}
		}()

		<-done
		log.Print("Stopping Folien server")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer func() { cancel() }()
		if err := s.Shutdown(ctx); err != nil {
			return err
		}
		return err
	},
}

func init() {
	serveCmd.Flags().StringVar(&keyPath, "keyPath", "folien", "Server private key path")
	serveCmd.Flags().StringVar(&host, "host", "localhost", "Server host to bind to")
	serveCmd.Flags().IntVar(&port, "port", 53531, "Server port to bind to")
}
