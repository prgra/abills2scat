package scat

import (
	"encoding/json"
	"fmt"
)

// TariffRootClass is a struct that represents a root class of a tariff
type TariffRootClass struct {
	Class int    `json:"class"`
	Rate  string `json:"rate"`
	Ceil  string `json:"ceil"`
}

// TariffProfile is a struct that represents a tariff profile
type TariffProfile struct {
	Name     string         `json:"-"`
	Type     string         `json:"type"`
	Outbound TariffSubClass `json:"outbound"`
	Inbound  TariffSubClass `json:"inbound"`
}

// TariffSubClass is a struct that represents a subclass of a tariff
type TariffSubClass struct {
	RootRate string            `json:"root_rate"`
	RootCeil string            `json:"root_ceil"`
	Classes  []TariffRootClass `json:"classes"`
}

type AbillsTariff struct {
	ID       int    `db:"tp_id"`
	InSpeed  int    `db:"in_speed"`
	OutSpeed int    `db:"out_speed"`
	Name     string `db:"name"`
}

func (n *Nas) SetTariffProfile(tp []TariffProfile) error {
	for i := range tp {
		b, _ := json.Marshal(tp[i])
		rc := fmt.Sprintf("fdpi_ctrl load profile --policing --profile.name %q --profile.json '%s'", tp[i].Name, string(b))
		out, err := n.Run(rc)
		if err != nil {
			return err
		}
		fmt.Println(out)
		// time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (a *App) GetTarifsFromAbills() (tarifs []TariffProfile, err error) {
	var tps []AbillsTariff
	err = a.Abills.DB.Select(&tps,
		`	SELECT
		tp.id as tp_id,
		tp.name as name,
		tt.in_speed,
		tt.out_speed
	FROM tarif_plans tp
	JOIN intervals i on (tp.tp_id = i.tp_id)
	JOIN trafic_tarifs tt on (tt.interval_id=i.id)
	WHERE tt.in_speed>0 and HOUR(CURTIME()) >= HOUR(i.begin) and HOUR(CURTIME()) <HOUR(i.end)`)
	for i := range tps {
		tarifs = append(tarifs, TariffProfile{
			Type: "HTB",
			Name: fmt.Sprintf("tp.%d", tps[i].ID),
			Outbound: TariffSubClass{
				RootRate: fmt.Sprintf("%dmbit", tps[i].OutSpeed/1000),
				RootCeil: fmt.Sprintf("%dmbit", tps[i].OutSpeed/1000),
				Classes: []TariffRootClass{
					{
						Class: 0,
						Rate:  "8bit",
						Ceil:  fmt.Sprintf("%dmbit", tps[i].OutSpeed/1000),
					},
					{
						Class: 1,
						Rate:  "8bit",
						Ceil:  fmt.Sprintf("%dmbit", tps[i].OutSpeed/1000),
					},
				},
			},
			Inbound: TariffSubClass{
				RootRate: fmt.Sprintf("%dmbit", tps[i].InSpeed/1000),
				RootCeil: fmt.Sprintf("%dmbit", tps[i].InSpeed/1000),
				Classes: []TariffRootClass{
					{
						Class: 0,
						Rate:  "8bit",
						Ceil:  fmt.Sprintf("%dmbit", tps[i].InSpeed/1000),
					},
					{
						Class: 1,
						Rate:  "8bit",
						Ceil:  fmt.Sprintf("%dmbit", tps[i].InSpeed/1000),
					},
				},
			},
		})
		for c := 2; c < 8; c++ {
			tarifs[i].Outbound.Classes = append(tarifs[i].Outbound.Classes, TariffRootClass{
				Class: c,
				Rate:  "8bit",
				Ceil:  fmt.Sprintf("%dmbit", tps[i].OutSpeed/1000),
			})
			tarifs[i].Inbound.Classes = append(tarifs[i].Inbound.Classes, TariffRootClass{
				Class: c,
				Rate:  "8bit",
				Ceil:  fmt.Sprintf("%dmbit", tps[i].InSpeed/1000),
			})
		}
	}
	return tarifs, err
}
