package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/anthdm/hollywood/actor"
)

type Utente struct {
	Username string
}

func NewUtente(username string) actor.Producer {
	return func() actor.Receiver {
		return &Utente{Username: username}
	}
}

func (u *Utente) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case string:
		fmt.Printf("[%s]: %s\n", u.Username, msg)
	}
}

type Controller struct {
	Pids []*actor.PID
}

func NewController() actor.Producer {
	return func() actor.Receiver {
		return &Controller{Pids: make([]*actor.PID, 0)}
	}
}

func (c *Controller) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case string:
		if len(c.Pids) >= 2 {
			ctx.Respond("Limite utenti raggiunto!")
			return
		}
		if msg == "NEGRO" {
			ctx.Respond("Nome offensivo, riprova.")
			return
		}
		pid := ctx.SpawnChild(NewUtente(msg), msg)
		c.Pids = append(c.Pids, pid)
		ctx.Respond("OK")
	case int:
		if msg >= 0 && msg < len(c.Pids) {
			ctx.Respond(c.Pids[msg])
		}
	}
}

func main() {
	e, _ := actor.NewEngine(actor.NewEngineConfig())
	controller := e.Spawn(NewController(), "controller")

	scanner := bufio.NewScanner(os.Stdin)

	// Registrazione utenti
	for lenRequest := 0; lenRequest < 2; {
		fmt.Printf("Vuoi unirti alla chat? (YES/NO): ")
		scanner.Scan()
		if scanner.Text() == "YES" {
			fmt.Printf("Inserisci Username: ")
			scanner.Scan()
			username := scanner.Text()

			resp := e.Request(controller, username, time.Second)
			if resp != nil {
				res, _ := resp.Result()
				if res == "OK" {
					lenRequest++
					fmt.Printf("Utente %s registrato!\n", username)
				} else {
					fmt.Println(res)
				}
			}
		}
	}

	fmt.Println("✅ Chat pronta. Ora potete parlare a turni!")

	// Conversazione a turni
	for {
		for turno := 0; turno < 2; turno++ {
			req := e.Request(controller, turno, time.Millisecond*200)
			if req == nil {
				fmt.Println("Errore richiesta PID")
				continue
			}
			msg, _ := req.Result()
			pid := msg.(*actor.PID)

			fmt.Printf("Turno Utente %d: scrivi messaggio >> ", turno+1)
			scanner.Scan()
			testo := scanner.Text()

			// invia il messaggio all’altro utente
			destIndex := 1 - turno
			destReq := e.Request(controller, destIndex, time.Millisecond*200)
			if destReq != nil {
				dest, _ := destReq.Result()
				e.Send(dest.(*actor.PID), testo)
			}
			_ = pid
		}
	}

}
