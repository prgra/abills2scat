package scat

import (
	"database/sql"
)

type AbillsUser struct {
	UID      string          `db:"uid"`
	TPID     int             `db:"tp_id"`
	IP       string          `db:"ip"`
	SpeedIn  int             `db:"in_speed"`
	SpeedOut int             `db:"out_speed"`
	Deposit  float32         `db:"deposit"`
	CCredit  sql.NullFloat64 `db:"ccredit"`
	UCredit  float64         `db:"ucredit"`
	CID      string          `db:"cid"`
	CalcInet bool            `db:"-"`
}

func (a App) GetUserList() ([]AbillsUser, error) {
	var users []AbillsUser
	err := a.Abills.DB.Select(&users, `	SELECT
		u.uid,
		INET_NTOA(dv.ip) as ip,
		tt.in_speed,
		tt.out_speed,
		tp.id as tp_id,
		if(u.company_id > 0, cb.deposit, b.deposit) AS deposit,
		c.credit as ccredit,
		u.credit as ucredit,
		dv.cid
	FROM
		users u
	JOIN dv_main dv on dv.uid=u.uid
	JOIN tarif_plans tp on (dv.tp_id=tp.id)
	JOIN intervals i on (tp.tp_id = i.tp_id)
	JOIN trafic_tarifs tt on (tt.interval_id=i.id)
	left JOIN bills b on (b.uid=u.uid)
	left join companies c on c.id=u.company_id
	left join bills cb on c.bill_id=cb.id
	where u.disable+u.deleted=0 and dv.ip>0 and tt.in_speed>0 and HOUR(CURTIME()) >= HOUR(i.begin) and HOUR(CURTIME()) <HOUR(i.end)`)
	if err != nil {
		return nil, err
	}
	for i := range users {
		if users[i].CCredit.Valid && users[i].UCredit == 0 {
			users[i].UCredit = users[i].CCredit.Float64
		}
		if users[i].UCredit+float64(users[i].Deposit) >= 0 {
			users[i].CalcInet = true
		}
	}
	return users, nil
}
