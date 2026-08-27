package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zyvorai/chimera/internal/config"
	"github.com/zyvorai/chimera/internal/lab"
	"github.com/zyvorai/chimera/internal/selftest"
)

func main() {
	if len(os.Args) < 2 {
		serve(os.Args[1:])
		return
	}
	switch os.Args[1] {
	case "serve":
		serve(os.Args[2:])
	case "selftest":
		runSelftest(os.Args[2:])
	case "print-config":
		printConfig(os.Args[2:])
	case "version":
		fmt.Println("chimera dev")
	default:
		serve(os.Args[1:])
	}
}

func serve(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	configPath := fs.String("config", "", "JSON config file")
	listen := fs.String("listen", "", "override listen address")
	_ = fs.Parse(args)
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	if *listen != "" {
		cfg.Listen = *listen
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	l, err := lab.Start(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Chimera ready\n")
	fmt.Printf("  endpoint: %s\n", l.URL.String())
	fmt.Printf("  username: %s\n", cfg.Username)
	fmt.Printf("  password: %s\n", cfg.Password)
	fmt.Printf("  admin:    %s://%s/__chimera/\n", l.URL.Scheme, l.URL.Host)
	fmt.Printf("  login:    %s / %s\n", cfg.AdminUsername, cfg.AdminPassword)
	if cfg.AdminPasswordFile != "" {
		fmt.Printf("            (generated; persisted at %s, reused on restart)\n", cfg.AdminPasswordFile)
	}
	fmt.Printf("  token:    %s\n", cfg.AdminToken)
	fmt.Printf("  sample VM path: /DC0/vm/DC0_C0_RP0_VM0\n")
	<-ctx.Done()
	cctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = l.Close(cctx)
}

func runSelftest(args []string) {
	fs := flag.NewFlagSet("selftest", flag.ExitOnError)
	endpoint := fs.String("url", "http://127.0.0.1:8989/sdk", "vCenter URL")
	user := fs.String("user", "administrator@vsphere.local", "username")
	pass := fs.String("pass", "vmware", "password")
	insecure := fs.Bool("insecure", true, "skip TLS verification")
	vm := fs.String("vm", "", "target a specific VM by name instead of the first one found")
	save := fs.String("save", "", "write the full exported disk to this path instead of only reading 4KB")
	_ = fs.Parse(args)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	r, err := selftest.Run(ctx, *endpoint, *user, *pass, *insecure, *vm, *save)
	if err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	fmt.Println(string(b))
}

func printConfig(args []string) {
	fs := flag.NewFlagSet("print-config", flag.ExitOnError)
	path := fs.String("config", "", "config file")
	_ = fs.Parse(args)
	cfg, err := config.Load(*path)
	if err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println(string(b))
}
