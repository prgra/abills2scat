package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/alexflint/go-arg"
	"github.com/prgra/abills2scat/scat"
)

func main() {
	var conf scat.Config
	_, err := toml.DecodeFile("scat.toml", &conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var args struct {
		ProfileSync bool `arg:"-t,--synctp" usage:"sync tarif profiles"`
	}
	arg.MustParse(&args)

	app, err := scat.NewApp(conf)
	if err != nil {
		fmt.Println("newapp", err)
		os.Exit(1)
	}
	if args.ProfileSync {
		tps, err := app.GetTarifsFromAbills()
		if err != nil {
			log.Fatal("Failed to parse output: " + err.Error())
		}
		for _, n := range app.Nases {
			err = n.SetTariffProfile(tps)
			if err != nil {
				log.Fatal("Failed to parse output: " + err.Error())
			}
		}
	}
	for i := range app.Nases {
		usrs, err := app.Nases[i].GetUserList()
		if err != nil {
			log.Fatal("Failed to parse output: " + err.Error())
		}

		usermap := make(map[string]scat.ShortScatUser)

		for _, u := range usrs {
			usermap[u.ID] = u
		}
		prf, err := app.Nases[i].GetUserProfilesList()
		if err != nil {
			log.Fatal("Failed to parse output: " + err.Error())
		}
		prfmap := make(map[string]scat.ScatUser)
		for i := range prf {
			uid := strings.Replace(prf[i].ID, "UID.", "", 1)
			u, ok := usermap[uid]
			if ok {
				prf[i].IP = u.IP
			}
			prfmap[prf[i].ID] = prf[i]
		}

		ausrs, uerr := app.GetUserList()

		if uerr != nil {
			log.Fatal("Failed to get users: " + uerr.Error())
		}

		aumap := make(map[string]scat.AbillsUser)

		for _, u := range ausrs {
			if u.CalcInet {
				aumap[u.UID] = u
			}
		}

		for id, u := range aumap {
			if _, ok := usermap[id]; !ok {
				app.Nases[i].Run(fmt.Sprintf("fdpi_ctrl load --bind_multi --user UID.%s:%s", u.UID, u.IP))
				_ = id
			}
			// fmt.Println("DEBUG", id, prfmap[fmt.Sprintf("UID.%s", id)], prfmap[fmt.Sprintf("UID.%s", id)].TPName, fmt.Sprintf("tssp.%d", u.TPID))
			pr, ok := prfmap[fmt.Sprintf("UID.%s", id)]
			if (ok && pr.TPName != fmt.Sprintf("tp.%d", u.TPID)) ||
				(ok && pr.IP != u.IP) ||
				!ok {
				_, err = app.Nases[i].Run(fmt.Sprintf("fdpi_ctrl load --policing --profile.name tp.%d --login UID.%s", u.TPID, u.UID))
				if err != nil {
					log.Print("Failed to parse output: " + err.Error())
				}
			}
		}
		for id, u := range usermap {
			if _, ok := aumap[id]; !ok {
				app.Nases[i].Run(fmt.Sprintf("fdpi_ctrl del --bind_multi --ip %s", u.IP))
				app.Nases[i].Run(fmt.Sprintf("fdpi_ctrl del --bind_multi --login %s", u.ID))
			}
		}
	}
}
