package discord

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/V4NSH4J/discord-mass-dm-GO/instance"
	"github.com/V4NSH4J/discord-mass-dm-GO/utilities"
	"github.com/zenthangplus/goccm"
)

func LaunchinviteJoiner() {
	utilities.PrintMenu([]string{"Convite Único", "Vários Convites (de Arquivo)"})
	invitechoice := utilities.UserInputInteger("Digite sua escolha:")
	if invitechoice != 1 && invitechoice != 2 {
		utilities.LogErr("Opção inválida")
		return
	}
	switch invitechoice {
	case 1:
		cfg, instances, err := instance.GetEverything()
		if err != nil {
			utilities.LogErr("Erro ao carregar configurações ou instâncias: %s", err)
		}
		var tokenFile, jointFile, failedFile, reactedFile string
		if cfg.OtherSettings.Logs {
			path := fmt.Sprintf(`logs/convite_unico/LOGS-CONVITE-%s-%s`, time.Now().Format(`2006-01-02 15-04-05`), utilities.RandStringBytes(5))
			err := os.MkdirAll(path, 0755)
			if err != nil && !os.IsExist(err) {
				utilities.LogErr("Erro ao criar o diretório de logs: %s", err)
				utilities.ExitSafely()
			}
			tokenFileX, err := os.Create(fmt.Sprintf(`%s/token.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de token: %s", err)
				utilities.ExitSafely()
			}
			tokenFileX.Close()
			jointFileX, err := os.Create(fmt.Sprintf(`%s/sucesso.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de sucesso: %s", err)
				utilities.ExitSafely()
			}
			jointFileX.Close()
			failedFileX, err := os.Create(fmt.Sprintf(`%s/falha.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de falha: %s", err)
				utilities.ExitSafely()
			}
			failedFileX.Close()
			reactedFileX, err := os.Create(fmt.Sprintf(`%s/reagiu.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de reação: %s", err)
				utilities.ExitSafely()
			}
			reactedFileX.Close()
			tokenFile, jointFile, failedFile, reactedFile = tokenFileX.Name(), jointFileX.Name(), failedFileX.Name(), reactedFileX.Name()
			for i := 0; i < len(instances); i++ {
				instances[i].WriteInstanceToFile(tokenFile)
			}
		}

		invite := utilities.UserInput("Digite o Código ou Link do Convite:")
		invite = processInvite(invite)
		threads := utilities.UserInputInteger("Digite o número de threads (0 para o máximo):")

		if threads > len(instances) {
			threads = len(instances)
		}
		if threads == 0 {
			threads = len(instances)
		}
		verif := utilities.UserInputInteger("Usar verificação adicional por reação? 0) Não 1) Sim")
		var channelid string
		var msgid string
		var emoji string

		if verif == 1 {
			channelid = utilities.UserInput("ID do canal com a mensagem de verificação:")
			msgid = utilities.UserInput("ID da mensagem com a reação de verificação:")
			emoji = utilities.UserInput("Digite o emoji:")
		}
		base := utilities.UserInputInteger("Digite o atraso base por thread (em segundos): ")
		random := utilities.UserInputInteger("Digite o atraso aleatório adicional por thread (em segundos): ")
		var delay int
		if random > 0 {
			delay = base + rand.Intn(random)
		} else {
			delay = base
		}
		c := goccm.New(threads)
		for i := 0; i < len(instances); i++ {
			c.Wait()
			go func(i int) {
				err := instances[i].Invite(invite)
				if err != nil {
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(failedFile, instances[i].Token)
					}
				} else {
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(jointFile, instances[i].Token)
					}
				}
				if verif == 1 {
					err := instances[i].React(channelid, msgid, emoji)
					if err != nil {
						utilities.LogFailed("%v falhou ao reagir com %v", instances[i].CensorToken(), emoji)
					} else {
						utilities.LogSuccess("%v reagiu com o emoji %v", instances[i].CensorToken(), emoji)
						if cfg.OtherSettings.Logs {
							utilities.WriteLinesPath(reactedFile, instances[i].Token)
						}
					}
				}
				time.Sleep(time.Duration(delay) * time.Second)
				c.Done()

			}(i)
		}
		c.WaitAllDone()
		utilities.LogSuccess("Todas as Threads Concluídas!")

	case 2:
		cfg, instances, err := instance.GetEverything()
		if err != nil {
			utilities.LogErr("Erro ao carregar configurações ou instâncias: %s", err)
		}
		var tokenFile string
		path := fmt.Sprintf(`logs/convite_multiplos/LOGS-CONVITES-%s-%s`, time.Now().Format(`2006-01-02 15-04-05`), utilities.RandStringBytes(5))
		if cfg.OtherSettings.Logs {
			err := os.MkdirAll(path, 0755)
			if err != nil && !os.IsExist(err) {
				utilities.LogErr("Erro ao criar o diretório de logs: %s", err)
				utilities.ExitSafely()
			}
			tokenFileX, err := os.Create(fmt.Sprintf(`%s/token.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de token: %s", err)
				utilities.ExitSafely()
			}
			tokenFileX.Close()
			tokenFile = tokenFileX.Name()
			for i := 0; i < len(instances); i++ {
				instances[i].WriteInstanceToFile(tokenFile)
			}
		}
		invites, err := utilities.ReadLines("invite.txt")
		if err != nil {
			utilities.LogErr("Erro ao abrir o arquivo invite.txt: %s", err)
			return
		}
		var inviteFiles []string
		if cfg.OtherSettings.Logs {
			for i := 0; i < len(invites); i++ {
				f, err := os.Create(fmt.Sprintf(`%s/%s.txt`, path, processInvite(invites[i])))
				if err != nil {
					utilities.LogErr("Erro ao criar o arquivo de log para o convite %v: %s", invites[i], err)
				}
				inviteFiles = append(inviteFiles, f.Name())
			}
		}

		if len(invites) == 0 {
			utilities.LogErr("Nenhum convite encontrado em invite.txt")
			return
		}
		delay := utilities.UserInputInteger("Digite o atraso entre entradas consecutivas por token (em segundos): ")
		threads := utilities.UserInputInteger("Digite o número de threads (0 para o máximo):")
		if threads > len(instances) {
			threads = len(instances)
		}
		if threads == 0 {
			threads = len(instances)
		}
		c := goccm.New(threads)
		for i := 0; i < len(instances); i++ {
			time.Sleep(time.Duration(cfg.DirectMessage.Offset) * time.Millisecond)
			c.Wait()
			go func(i int) {
				for j := 0; j < len(invites); j++ {
					err := instances[i].Invite(processInvite(invites[j]))
					if err == nil {
						if cfg.OtherSettings.Logs {
							utilities.WriteLinesPath(inviteFiles[j], instances[i].Token)
						}
					}
					time.Sleep(time.Duration(delay) * time.Second)
				}
				c.Done()
			}(i)
		}
		c.WaitAllDone()
		utilities.LogSuccess("Todas as Threads Concluídas!")
	}
}

func processInvite(rawInvite string) string {
	if !strings.Contains(rawInvite, "/") {
		return rawInvite
	} else {
		return strings.Split(rawInvite, "/")[len(strings.Split(rawInvite, "/"))-1]
	}
}