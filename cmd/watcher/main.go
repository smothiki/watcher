package main

import (
	"fmt"
	"os"
	"time"

	"github.com/srelab/common/log"

	"github.com/srelab/watcher/pkg"
	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/util"

	"github.com/urfave/cli"
)

func main() {
	app := &cli.App{
		Name:     g.NAME,
		Usage:    "Kubernetes Watch Service",
		Version:  g.VERSION,
		Compiled: time.Now(),
		Authors:  []cli.Author{{Name: g.AUTHOR, Email: g.MAIL}},
		Before: func(c *cli.Context) error {
			fmt.Fprintf(c.App.Writer, util.StripIndent(
				`
				#    #   ##   #####  ####  #    # ###### #####
				#    #  #  #    #   #    # #    # #      #    # 
				#    # #    #   #   #      ###### #####  #    # 
				# ## # ######   #   #      #    # #      #####  
				##  ## #    #   #   #    # #    # #      #   #  
				#    # #    #   #    ####  #    # ###### #    #  
			`))
			return nil
		},
		Commands: []cli.Command{
			{
				Name:  "start",
				Usage: "Start a new watcher",
				Action: func(ctx *cli.Context) {
					for _, flagName := range ctx.FlagNames() {
						if ctx.String(flagName) != "" {
							continue
						}

						fmt.Println(flagName + " is required")
						os.Exit(127)
					}

					if err := g.ReadInConfig(ctx); err != nil {
						panic(fmt.Sprintf("could not read config: %v", err))
					}

					log.Init(g.Config().Log)
					pkg.Start()
				},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config_file, cf", Usage: "Load configuration from `FILE`"},
				},
			},
		},
	}

	app.Run(os.Args)
}
