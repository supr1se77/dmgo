package discord

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/V4NSH4J/discord-mass-dm-GO/instance"
	"github.com/V4NSH4J/discord-mass-dm-GO/utilities"
)

// Inicia o modo de DM em massa com evasão de detecção.
func LaunchAntiAntiRaidMode() {
	// Carregando tudo o que é necessário
	cfg, instances, err := instance.GetEverything()
	if err != nil {
		utilities.LogErr("Erro ao carregar configurações ou instâncias: %s", err)
		return
	}
	memberids, err := utilities.ReadLines("memberids.txt")
	if err != nil {
		utilities.LogErr("Erro ao ler o arquivo memberids.txt: %s", err)
		return
	}
	if cfg.DirectMessage.Skip {
		completed, err := utilities.ReadLines("completed.txt")
		if err != nil {
			utilities.LogErr("Erro ao abrir o arquivo completed.txt: %s", err)
			return
		}
		memberids = utilities.RemoveSubset(memberids, completed)
	}
	if cfg.DirectMessage.SkipFailed {
		failedSkip, err := utilities.ReadLines("failed.txt")
		if err != nil {
			utilities.LogErr("Erro ao abrir o arquivo failed.txt: %s", err)
			return
		}
		memberids = utilities.RemoveSubset(memberids, failedSkip)
	}
	memberChan := make(chan string, len(memberids))
	for i := 0; i < len(memberids); i++ {
		go func(i int) {
			memberChan <- memberids[i]
		}(i)
	}
	threads := utilities.UserInputInteger("Digite o número de threads (Recomendado: <10):")
	serverid := utilities.UserInput("Digite o ID do servidor:")
	delayBetweenJoins := utilities.UserInputInteger("Digite o atraso entre as entradas no servidor (em segundos):")
	randomDelayBetweenJoins := utilities.UserInputInteger("Digite o atraso aleatório adicional entre as entradas:")
	invite := utilities.UserInput("Digite o convite:")
	var msg instance.Message
	messagechoice := utilities.UserInputInteger("Digite 1 para usar a mensagem do arquivo ou 2 para digitar no console: ")
	if messagechoice != 1 && messagechoice != 2 {
		utilities.LogErr("Opção inválida")
		return
	}
	if messagechoice == 2 {
		text := utilities.UserInput("Digite sua mensagem, use \\n para pular linha. Você também pode definir uma mensagem padrão em message.json")
		msg.Content = text
		msg.Content = strings.Replace(msg.Content, "\\n", "\n", -1)
		var msgs []instance.Message
		msgs = append(msgs, msg)
		err := instance.SetMessages(instances, msgs)
		if err != nil {
			utilities.LogErr("Erro ao definir as mensagens: %s", err)
			return
		}
	} else {
		var msgs []instance.Message
		err := instance.SetMessages(instances, msgs)
		if err != nil {
			utilities.LogErr("Erro ao definir as mensagens: %s", err)
			return
		}
	}
	ticker := make(chan bool)
	go func() {
		for {
			if randomDelayBetweenJoins == 0 && delayBetweenJoins != 0 {
				time.Sleep(time.Duration(delayBetweenJoins) * time.Second)
			} else if randomDelayBetweenJoins != 0 && delayBetweenJoins == 0 {
				time.Sleep(time.Duration(rand.Intn(randomDelayBetweenJoins)) * time.Second)
			} else {
				time.Sleep(time.Second * time.Duration(delayBetweenJoins+rand.Intn(randomDelayBetweenJoins)))
			}

			ticker <- true
		}
	}()
	var completed []string
	var failed []string
	var tokenUsed []instance.Instance
	dormantPool := make(chan Worker, len(instances))
	timedOutPool := make(chan Worker, threads)
	for i := 0; i < len(instances); i++ {
		go func(i int) {
			dormantPool <- Worker{
				Instance:     &instances[i],
				TimedOutTill: time.Now(),
				Valid:        false,
			}
		}(i)
	}
	tickerTitle := make(chan bool)
	go func() {
	Out:
		for {
			select {
			case <-tickerTitle:
				break Out
			default:
				var avg int
				if len(tokenUsed) == 0 {
					avg = 0
				} else {
					avg = len(completed) / len(tokenUsed)
				}
				// Título da janela do console modificado
				cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf(`SUPR1SE [%d enviadas, %d falhas, %d tokens não usados, %d tokens em espera, %d tokens usados, %d DMs em média]`, len(completed), len(failed), len(dormantPool), len(timedOutPool), len(tokenUsed), avg))
				_ = cmd.Run()
			}

		}
	}()
	var wg sync.WaitGroup
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
		Token:
			for {
				if len(timedOutPool) == 0 && len(dormantPool) == 0 {
					break Token
				}
				var w Worker
				// Prioridade para workers em espera
				w = AvailableWorker(timedOutPool)
				if w == (Worker{}) {
					// Tenta pegar do grupo inativo se não houver workers em espera
					if len(dormantPool) != 0 {
						w = NewWorker(dormantPool)
						if w == (Worker{}) {
							continue Token
						}
					} else if len(timedOutPool) != 0 {
						for {
							if len(timedOutPool) == 0 {
								break Token
							}
							w = AvailableWorker(timedOutPool)
							if w != (Worker{}) {
								break
							} else {
								continue
							}
						}
					}

				}
				tokenUsed = append(tokenUsed, *w.Instance)
			MassDM:
				for {
					if len(memberChan) == 0 {
						break MassDM
					}
					uuid := <-memberChan
				TokenCheck:
					for x := 0; x < 3; x++ {
						r := w.Instance.CheckToken()
						if r == 200 || r == 204 {
							w.Valid = true
							break TokenCheck
						} else if r == 403 || r == 401 {
							utilities.LogLocked("Token está bloqueado ou inválido: %s : %d", w.Instance.CensorToken(), r)
							continue Token
						} else {
							continue TokenCheck
						}
					}
					if !w.Valid {
						continue Token
					}
					var inServer bool
				ServerCheck:
					for x := 0; x < 3; x++ {
						r, err := w.Instance.ServerCheck(serverid)
						if err != nil {
							utilities.LogErr("Token %s: erro ao verificar se está no servidor %s : %s", w.Instance.CensorToken(), serverid, err)
							continue ServerCheck
						}
						if r == 200 || r == 204 {
							inServer = true
							break ServerCheck
						} else if r == 429 || r >= 500 {
							time.Sleep(time.Second * 5)
							continue ServerCheck
						} else {
							// Token não está no servidor
						TimeCheck:
							for {
								select {
								case <-ticker:
									err := w.Instance.Invite(invite)
									if err != nil {
										utilities.LogErr("Token %s: erro ao entrar no servidor %s : %s", w.Instance.CensorToken(), serverid, err)
										continue Token
									} else {
										inServer = true
										break ServerCheck
									}
								default:
									continue TimeCheck
								}
							}
						}
					}
					if !inServer {
						continue Token
					}
					snowflake, err := w.Instance.OpenChannel(uuid)
					if err != nil {
						utilities.LogErr("Token %s: erro ao abrir canal com %s : %s", w.Instance.CensorToken(), uuid, err)
						if w.Instance.Quarantined {
							continue Token
						}
						continue MassDM
					}
					r, bytes, err := w.Instance.SendMessage(snowflake, uuid)
					if err != nil {
						utilities.LogErr("Token %s: erro ao enviar mensagem para %s : %s", w.Instance.CensorToken(), uuid, err)
						continue MassDM
					}
					var resp jsonResponse
					err = json.Unmarshal(bytes, &resp)
					if err != nil {
						utilities.LogErr("Token %s: erro ao decodificar resposta de %s : %s", w.Instance.CensorToken(), uuid, err)
						continue MassDM
					}
					if r == 200 {
						utilities.LogSuccess("[%d] Token %s enviou mensagem para %s", len(completed), w.Instance.CensorToken(), uuid)
						err := utilities.WriteLine("input/completed.txt", uuid)
						if err != nil {
							utilities.LogErr("Erro ao escrever em completed.txt : %s", err)
						}
						completed = append(completed, uuid)
					} else if resp.Code == 20026 {
						utilities.LogFailed("Token %s está em quarentena", w.Instance.CensorToken())
						memberChan <- uuid
					} else if r == 403 && resp.Code == 40003 {
						utilities.LogInfo("Token %s atingiu o limite de taxa e entrará em espera.", w.Instance.CensorToken())
						w.TimedOutTill = time.Now().Add(time.Second * time.Duration(cfg.DirectMessage.LongDelay+rand.Intn(cfg.SuspicionAvoidance.RandomRateLimitDelay)))
						timedOutPool <- w
						memberChan <- uuid
						continue Token
					} else if r == 403 && resp.Code == 50007 {
						utilities.LogFailed("Token %s falhou ao enviar DM para %s (DMs fechadas ou sem servidores em comum)", w.Instance.CensorToken(), uuid)
						err := utilities.WriteLine("input/failed.txt", uuid)
						if err != nil {
							utilities.LogErr("Erro ao escrever em failed.txt : %s", err)
						}
						failed = append(failed, uuid)
					} else if (r == 403 && resp.Code == 40002) || r == 401 || r == 405 {
						utilities.LogLocked("Token %s está bloqueado", w.Instance.CensorToken())
						memberChan <- uuid
						continue Token
					} else if r == 403 && resp.Code == 50009 {
						utilities.LogFailed("Token %s falhou ao enviar DM para %s (nível de verificação baixo)", w.Instance.CensorToken(), uuid)
						err := utilities.WriteLine("input/failed.txt", uuid)
						if err != nil {
							utilities.LogErr("Erro ao escrever em failed.txt : %s", err)
						}
						failed = append(failed, uuid)
					} else if r == 429 {
						utilities.LogFailed("Token %s atingiu o limite de taxa", w.Instance.CensorToken())
						memberChan <- uuid
						time.Sleep(60 * time.Second)
					} else if r == 400 && strings.Contains(string(bytes), "captcha") {
						utilities.LogFailed("Token %s falhou ao enviar DM para %s (captcha resolvido incorretamente)", w.Instance.CensorToken(), uuid)
						memberChan <- uuid
						continue Token
					} else {
						utilities.LogFailed("Token %s falhou ao enviar DM para %s - Código de Status Desconhecido %d %s", w.Instance.CensorToken(), uuid, r, string(bytes))
						err := utilities.WriteLine("input/failed.txt", uuid)
						if err != nil {
							utilities.LogErr("Erro ao escrever em failed.txt : %s", err)
						}
						failed = append(failed, uuid)
						continue Token
					}
					if cfg.SuspicionAvoidance.RandomIndividualDelay != 0 {
						time.Sleep(time.Duration(cfg.DirectMessage.Delay+rand.Intn(cfg.SuspicionAvoidance.RandomIndividualDelay)) * time.Second)
					} else {
						time.Sleep(time.Duration(cfg.DirectMessage.Delay) * time.Second)
					}

					continue MassDM
				}
			}
		}(i)
	}
	wg.Wait()
	tickerTitle <- true
	utilities.LogSuccess("Envio de mensagens finalizado")

}

type Worker struct {
	Instance     *instance.Instance
	TimedOutTill time.Time
	Valid        bool
}

func (w *Worker) isTimedOut() bool {
	if time.Since(w.TimedOutTill) > 0 {
		return false
	} else {
		return true
	}
}

func AvailableWorker(timedOutPool chan Worker) Worker {
	if len(timedOutPool) == 0 {
		return Worker{}
	}
	for w := range timedOutPool {
		if !w.isTimedOut() {
			utilities.LogInfo("Reiniciando token em espera: %s", w.Instance.CensorToken())
			return w
		}
	}
	return Worker{}
}

func NewWorker(dormantPool chan Worker) Worker {
	if len(dormantPool) == 0 {
		return Worker{}
	}
	for w := range dormantPool {
		return w
	}
	return Worker{}
}