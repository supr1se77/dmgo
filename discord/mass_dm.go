package discord

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/V4NSH4J/discord-mass-dm-GO/instance"
	"github.com/V4NSH4J/discord-mass-dm-GO/utilities"
)

func LaunchMassDM() {
	members, err := utilities.ReadLines("memberids.txt")
	if err != nil {
		utilities.LogErr("Erro ao abrir o arquivo de IDs de Membros: %s", err)
		return
	}
	cfg, instances, err := instance.GetEverything()
	if err != nil {
		utilities.LogErr("Erro ao carregar configurações ou instâncias: %s", err)
		return
	}
	var tokenFile, completedUsersFile, failedUsersFile, lockedFile, quarantinedFile, logsFile string
	if cfg.OtherSettings.Logs {
		path := fmt.Sprintf(`logs/dm_em_massa/LOGS-DM-%s-%s`, time.Now().Format(`2006-01-02 15-04-05`), utilities.RandStringBytes(5))
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
		completedUsersFileX, err := os.Create(fmt.Sprintf(`%s/sucesso.txt`, path))
		if err != nil {
			utilities.LogErr("Erro ao criar arquivo de sucesso: %s", err)
			utilities.ExitSafely()
		}
		completedUsersFileX.Close()
		failedUsersFileX, err := os.Create(fmt.Sprintf(`%s/falha.txt`, path))
		if err != nil {
			utilities.LogErr("Erro ao criar arquivo de falha: %s", err)
			utilities.ExitSafely()
		}
		failedUsersFileX.Close()
		lockedFileX, err := os.Create(fmt.Sprintf(`%s/bloqueados.txt`, path))
		if err != nil {
			utilities.LogErr("Erro ao criar arquivo de bloqueados: %s", err)
			utilities.ExitSafely()
		}
		lockedFileX.Close()
		quarantinedFileX, err := os.Create(fmt.Sprintf(`%s/quarentena.txt`, path))
		if err != nil {
			utilities.LogErr("Erro ao criar arquivo de quarentena: %s", err)
			utilities.ExitSafely()
		}
		quarantinedFileX.Close()
		LogsX, err := os.Create(fmt.Sprintf(`%s/logs.txt`, path))
		if err != nil {
			utilities.LogErr("Erro ao criar arquivo de logs: %s", err)
			utilities.ExitSafely()
		}
		LogsX.Close()
		tokenFile, completedUsersFile, failedUsersFile, lockedFile, quarantinedFile, logsFile = tokenFileX.Name(), completedUsersFileX.Name(), failedUsersFileX.Name(), lockedFileX.Name(), quarantinedFileX.Name(), LogsX.Name()
		for i := 0; i < len(instances); i++ {
			instances[i].WriteInstanceToFile(tokenFile)
		}
	}
	if cfg.OtherSettings.Logs {
		utilities.WriteLinesPath(logsFile, fmt.Sprintf("Hora de Início: %v", time.Now()))
	}
	var msg instance.Message
	messagechoice := utilities.UserInputInteger("Digite 1 para usar a mensagem do arquivo, 2 para digitar no console: ")
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
	if cfg.OtherSettings.Logs {
		if len(instances) > 0 {
			utilities.WriteLinesPath(logsFile, fmt.Sprintf("Mensagens Carregadas: %v", instances[0].Messages))
		}
	}
	advancedchoice := utilities.UserInputInteger("Deseja usar Configurações Avançadas? 0: Não, 1: Sim: ")

	var checkchoice int
	var serverid string
	var tryjoinchoice int
	var invite string
	var maxattempts int
	if advancedchoice != 0 && advancedchoice != 1 {
		utilities.LogErr("Opção inválida")
		return
	}
	if advancedchoice == 1 {
		checkchoice = utilities.UserInputInteger("Deseja verificar se o token ainda está no servidor antes de cada DM? [0: Não, 1: Sim]")
		if checkchoice != 0 && checkchoice != 1 {
			utilities.LogErr("Opção inválida")
			return
		}
		if checkchoice == 1 {
			serverid = utilities.UserInput("Digite o ID do servidor: ")
			tryjoinchoice = utilities.UserInputInteger("Deseja tentar entrar novamente no servidor se o token não estiver nele? [0: Não, 1: Sim]")
			if tryjoinchoice != 0 && tryjoinchoice != 1 {
				utilities.LogErr("Opção inválida")
				return
			}
			if tryjoinchoice == 1 {
				invite = utilities.UserInput("Digite um código de convite permanente:")
				maxattempts = utilities.UserInputInteger("Digite o máximo de tentativas para entrar novamente:")
			}
		}
	}
	// Inicia variáveis e slices para logs e contagem
	var session []string
	var completed []string
	var failed []string
	var dead []string
	var failedCount = 0
	var openedChannels = 0
	completed, err = utilities.ReadLines("completed.txt")
	if err != nil {
		utilities.LogErr("Erro ao abrir o arquivo completed.txt: %s", err)
		return
	}
	if cfg.DirectMessage.Skip {
		members = utilities.RemoveSubset(members, completed)
		if cfg.OtherSettings.Logs {
			utilities.WriteLinesPath(logsFile, fmt.Sprintf("Usuários na lista negra de completed.txt: %v", len(completed)))
		}
	}
	if cfg.DirectMessage.SkipFailed {
		failedSkip, err := utilities.ReadLines("failed.txt")
		if err != nil {
			utilities.LogErr("Erro ao abrir o arquivo failed.txt: %s", err)
			return
		}
		if cfg.OtherSettings.Logs {
			utilities.WriteLinesPath(logsFile, fmt.Sprintf("Usuários na lista negra de failed.txt: %v", len(failedSkip)))
		}
		members = utilities.RemoveSubset(members, failedSkip)
	}
	if len(instances) == 0 {
		utilities.LogErr("Coloque seus tokens em tokens.txt")
		if cfg.OtherSettings.Logs {
			utilities.WriteLinesPath(logsFile, fmt.Sprintf("Tokens carregados: %v", len(instances)))
		}
		return
	}
	if len(members) == 0 {
		utilities.LogErr("Coloque os IDs dos membros e garanta que não estejam todos em completed.txt ou failed.txt")
		return
	}
	if len(members) < len(instances) {
		instances = instances[:len(members)]
	}
	if cfg.OtherSettings.Logs {
		utilities.WriteLinesPath(logsFile, fmt.Sprintf("Membros únicos carregados: %v", len(members)))
	}
	msgs := instances[0].Messages
	for i := 0; i < len(msgs); i++ {
		if msgs[i].Content == "" && msgs[i].Embeds == nil {
			utilities.LogWarn("A mensagem %v está vazia", i)
		}
	}
	// Envia membros para um canal
	mem := make(chan string, len(members))
	go func() {
		for i := 0; i < len(members); i++ {
			mem <- members[i]
		}
	}()
	ticker := make(chan bool)
	// Define informações na barra de título da janela
	go func() {
	Out:
		for {
			select {
			case <-ticker:
				break Out
			default:
				cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf(`SUPR1SE [%d enviadas, %v falhas, %d bloq., %v média DMs, %v média canais, %d tokens rest.]`, len(session), len(failed), len(dead), len(session)/len(instances), openedChannels/len(instances), len(instances)-len(dead)))
				_ = cmd.Run()
			}
		}
	}()
	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < len(instances); i++ {
		// Atraso entre goroutines
		time.Sleep(time.Duration(cfg.DirectMessage.Offset) * time.Millisecond)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for {
				// Pega um membro do canal
				if len(mem) == 0 {
					break
				}
				member := <-mem
				instances[i].LastIDstr = ""
				// Para o loop se o máximo de DMs for atingido
				if cfg.DirectMessage.MaxDMS != 0 && instances[i].Count >= cfg.DirectMessage.MaxDMS {
					utilities.LogInfo("Máximo de DMs atingido para %v", instances[i].CensorToken())
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token atingiu o máximo de DMs", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken()))
					}
					break
				}
				// Inicia conexão websocket
				if cfg.DirectMessage.Websocket && instances[i].Ws == nil {
					err := instances[i].StartWS()
					if err != nil {
						utilities.LogFailed("Erro ao abrir o websocket: %v", err)
					} else {
						utilities.LogSuccess("Websocket aberto para %v", instances[i].CensorToken())
					}
				}
				// Verifica se o token é válido
				status := instances[i].CheckToken()
				if status != 200 && status != 204 && status != 429 && status != -1 {
					failedCount++
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(failedUsersFile, member)
					}
					utilities.LogLocked("Token %v pode estar bloqueado - Parando instância e adicionando membro à lista de falhas. %v [%v]", instances[i].CensorToken(), status, failedCount)
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token bloqueado", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken()))
					}
					failed = append(failed, member)
					dead = append(dead, instances[i].Token)
					if cfg.OtherSettings.Logs {
						instances[i].WriteInstanceToFile(lockedFile)
					}
					err := utilities.WriteLine("input/failed.txt", member)
					if err != nil {
						utilities.LogErr("Erro ao escrever em failed.txt: %s", err)
					}
					if cfg.DirectMessage.Stop {
						break
					}
				}
				// Opções Avançadas
				if advancedchoice == 1 {
					if checkchoice == 1 {
						r, err := instances[i].ServerCheck(serverid)
						if err != nil {
							utilities.LogErr("Erro ao verificar o servidor: %s", err)
							continue
						}
						if r != 200 && r != 204 && r != 429 {
							if tryjoinchoice == 0 {
								utilities.LogFailed("Token %s não está no servidor %s", instances[i].CensorToken(), serverid)
								if cfg.OtherSettings.Logs {
									utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token não está no servidor %v", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken(), serverid))
								}
								break
							} else {
								if instances[i].Retry >= maxattempts {
									if cfg.OtherSettings.Logs {
										utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token atingiu o máximo de tentativas para entrar", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken()))
									}
									utilities.LogFailed("Parando o token %v [Máximo de tentativas para entrar no servidor atingido]", instances[i].CensorToken())
									break
								}
								err := instances[i].Invite(invite)
								if err != nil {
									utilities.LogFailed("Erro ao entrar no servidor: %v", err)
									if cfg.OtherSettings.Logs {
										utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token com erro ao entrar no servidor %v", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken(), err))
									}
									instances[i].Retry++
									continue
								}
							}
						}
					}
				}
				var user string
				user = member
				// Verifica Servidores em Comum
				if cfg.DirectMessage.Mutual {
					info, err := instances[i].UserInfo(member)
					if err != nil {
						failedCount++
						if cfg.OtherSettings.Logs {
							utilities.WriteLinesPath(failedUsersFile, member)
						}
						utilities.LogErr("Erro ao obter informações do usuário: %v [%v]", err, failedCount)
						err = utilities.WriteLine("input/failed.txt", member)
						if err != nil {
							utilities.LogErr("Erro ao escrever em failed.txt: %s", err)
						}
						failed = append(failed, member)
						continue
					}
					if len(info.Mutual) == 0 {
						failedCount++
						if cfg.OtherSettings.Logs {
							utilities.WriteLinesPath(failedUsersFile, member)
						}
						utilities.LogFailed("Token %v falhou ao enviar DM para %v [Nenhum Servidor em Comum] [%v]", instances[i].CensorToken(), info.User.Username+info.User.Discriminator, failedCount)
						if cfg.OtherSettings.Logs {
							utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: falha ao enviar DM para %v [Sem servidores em comum]", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken(), member))
						}
						err = utilities.WriteLine("input/failed.txt", member)
						if err != nil {
							utilities.LogErr("Erro ao escrever em failed.txt: %s", err)
						}
						failed = append(failed, member)
						continue
					}
					user = info.User.Username + "#" + info.User.Discriminator
					// Usado apenas se Websocket estiver ativo
					if cfg.DirectMessage.Friend && cfg.DirectMessage.Websocket {
						x, err := strconv.Atoi(info.User.Discriminator)
						if err != nil {
							utilities.LogErr("Erro ao converter o discriminador para int: %v", err)
							continue
						}
						resp, err := instances[i].Friend(info.User.Username, x)
						if err != nil {
							utilities.LogErr("Erro ao enviar pedido de amizade: %v", err)
							continue
						}
						defer resp.Body.Close()
						if resp.StatusCode != 204 && err != nil {
							if !errors.Is(err, io.ErrUnexpectedEOF) {
								body, err := utilities.ReadBody(*resp)
								if err != nil {
									utilities.LogErr("Erro ao ler o corpo da resposta: %v", err)
									continue
								}
								utilities.LogFailed("Erro ao enviar pedido de amizade: %v", body)
								continue
							}
							utilities.LogErr("Erro ao enviar pedido de amizade: %v", err)
							continue
						} else {
							utilities.LogSuccess("Pedido de amizade enviado para %v", info.User.Username+info.User.Discriminator)
							if cfg.OtherSettings.Logs {
								utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token enviou amizade para %v", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken(), member))
							}
						}
					}
				}
				// Abre o canal para obter o snowflake
				snowflake, err := instances[i].OpenChannel(member)
				if err != nil {
					failedCount++
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(failedUsersFile, member)
					}
					utilities.LogErr("Erro ao abrir canal de DM: %v [%v]", err, failedCount)
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token com erro %v ao abrir canal", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken(), err))
					}
					err = utilities.WriteLine("input/failed.txt", member)
					if err != nil {
						utilities.LogErr("Erro ao escrever em failed.txt: %s", err)
					}
					failed = append(failed, member)
					if instances[i].Quarantined {
						break
					}
					continue
				}
				if cfg.SuspicionAvoidance.RandomDelayOpenChannel != 0 {
					time.Sleep(time.Duration(rand.Intn(cfg.SuspicionAvoidance.RandomDelayOpenChannel)) * time.Second)
				}
				respCode, body, err := instances[i].SendMessage(snowflake, member)
				openedChannels++
				if err != nil {
					failedCount++
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(failedUsersFile, member)
					}
					utilities.LogErr("Erro ao enviar mensagem: %v [%v]", err, failedCount)
					err = utilities.WriteLine("input/failed.txt", member)
					if err != nil {
						utilities.LogErr("Erro ao escrever em failed.txt: %s", err)
					}
					failed = append(failed, member)
					continue
				}
				var response jsonResponse
				errx := json.Unmarshal(body, &response)
				if errx != nil {
					failedCount++
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(failedUsersFile, member)
					}
					utilities.LogErr("Erro ao decodificar a resposta: %v [%v]", errx, failedCount)
					err = utilities.WriteLine("input/failed.txt", member)
					if err != nil {
						utilities.LogErr("Erro ao escrever em failed.txt: %s", err)
					}
					failed = append(failed, member)
					continue
				}
				// Tudo certo, continua
				if respCode == 200 {
					err = utilities.WriteLine("input/completed.txt", member)
					if err != nil {
						utilities.LogErr("Erro ao escrever em completed.txt: %s", err)
					}
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(completedUsersFile, member)
					}
					completed = append(completed, member)
					session = append(session, member)
					utilities.LogSuccess("[DM-%v] Token %v enviou DM para %v", len(session), instances[i].CensorToken(), user)
					if cfg.DirectMessage.Websocket && cfg.DirectMessage.Call && instances[i].Ws != nil {
						err := instances[i].Call(snowflake)
						if err != nil {
							utilities.LogErr("Token %v: Erro ao ligar para %v: %v", instances[i].CensorToken(), user, err)
						}
					}
					if cfg.DirectMessage.Block {
						r, err := instances[i].BlockUser(member)
						if err != nil {
							utilities.LogErr("Erro ao bloquear usuário: %v", err)
						} else {
							if r == 204 {
								utilities.LogSuccess("Bloqueado %v", user)
								if cfg.OtherSettings.Logs {
									utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token bloqueou o usuário %v", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken(), member))
								}
							} else {
								utilities.LogErr("Erro ao bloquear usuário: %v", r)
							}
						}
					}
					if cfg.DirectMessage.Close {
						r, err := instances[i].CloseDMS(snowflake)
						if err != nil {
							utilities.LogErr("Erro ao fechar DM: %v", err)
						} else {
							if r == 200 {
								utilities.LogSuccess("Fechado %v", user)
								if cfg.OtherSettings.Logs {
									utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token fechou a DM do usuário %v", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken(), member))
								}
							} else {
								utilities.LogErr("Erro ao fechar DM: %v", r)
							}
						}
					}
				} else if response.Code == 20026 {
					utilities.LogLocked("Token %v está em Quarentena, considerando como bloqueado", instances[i].CensorToken())
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token em quarentena", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken()))
					}
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(failedUsersFile, member)
					}
					dead = append(dead, instances[i].Token)
					if cfg.OtherSettings.Logs {
						instances[i].WriteInstanceToFile(lockedFile)
					}
					if cfg.DirectMessage.Stop {
						break
					}
					if cfg.OtherSettings.Logs {
						instances[i].WriteInstanceToFile(quarantinedFile)
						utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token em quarentena", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken()))
					}
					mem <- member
				} else if respCode == 403 && response.Code == 40003 {
					mem <- member
					utilities.LogInfo("Token %v em espera por %v minutos!", instances[i].CensorToken(), int(cfg.DirectMessage.LongDelay/60))
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token com limite de taxa, esperando por %v segundos", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken(), cfg.DirectMessage.LongDelay))
					}
					time.Sleep(time.Duration(cfg.DirectMessage.LongDelay) * time.Second)
					if cfg.SuspicionAvoidance.RandomRateLimitDelay != 0 {
						time.Sleep(time.Duration(rand.Intn(cfg.SuspicionAvoidance.RandomRateLimitDelay)) * time.Second)
					}
					utilities.LogInfo("Token %v continuando!", instances[i].CensorToken())
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token continuando", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken()))
					}
				} else if respCode == 403 && response.Code == 50007 {
					failedCount++
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(failedUsersFile, member)
					}
					failed = append(failed, member)
					err = utilities.WriteLine("input/failed.txt", member)
					if err != nil {
						utilities.LogErr("Erro ao escrever em failed.txt: %s", err)
					}
					utilities.LogFailed("Token %v falhou ao enviar DM para %v (DMs fechadas ou sem servidor em comum)", instances[i].CensorToken(), user)
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: falha ao enviar DM para %v [DMs fechadas ou sem servidor em comum]", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken(), member))
					}
				} else if (respCode == 403 && response.Code == 40002) || respCode == 401 || respCode == 405 {
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(failedUsersFile, member)
					}
					utilities.LogFailed("Token %v está bloqueado ou desativado. Parando instância. %v %v", instances[i].CensorToken(), respCode, string(body))
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token bloqueado ou desativado", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken()))
					}
					dead = append(dead, instances[i].Token)
					if cfg.OtherSettings.Logs {
						instances[i].WriteInstanceToFile(lockedFile)
					}
					if cfg.DirectMessage.Stop {
						break
					}
					mem <- member
				} else if respCode == 403 && response.Code == 50009 {
					failedCount++
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(failedUsersFile, member)
					}
					failed = append(failed, member)
					err = utilities.WriteLine("input/failed.txt", member)
					if err != nil {
						utilities.LogErr("Erro ao escrever em failed.txt: %s", err)
					}
					utilities.LogFailed("Token %v não pode enviar DM para %v. Pode não ter passado na triagem de membros ou o nível de verificação é muito baixo. %v", instances[i].CensorToken(), user, string(body))
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: nível de verificação do canal muito alto", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken()))
					}
				} else if respCode == 429 {
					utilities.LogFailed("Token %v está com limite de taxa. Esperando por 10 segundos", instances[i].CensorToken())
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: token com limite de taxa", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken()))
					}
					mem <- member
					time.Sleep(10 * time.Second)
				} else if respCode == 400 && strings.Contains(string(body), "captcha") {
					mem <- member
					utilities.LogFailed("Token %v: Captcha resolvido incorretamente", instances[i].CensorToken())
					if instances[i].Config.CaptchaSettings.CaptchaAPI == "anti-captcha.com" {
						err := instances[i].ReportIncorrectRecaptcha()
						if err != nil {
							utilities.LogFailed("Erro ao reportar hcaptcha incorreto: %v", err)
						} else {
							utilities.LogSuccess("Hcaptcha incorreto reportado com sucesso [%v]", instances[i].LastID)
						}
					}
					instances[i].Retry++
					if instances[i].Retry >= cfg.CaptchaSettings.MaxCaptchaDM && cfg.CaptchaSettings.MaxCaptchaDM != 0 {
						utilities.LogFailed("Parando o token %v (máximo de captchas resolvidos atingido)", instances[i].CensorToken())
						break
					}
				} else {
					failedCount++
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(failedUsersFile, member)
					}
					failed = append(failed, member)
					err = utilities.WriteLine("input/failed.txt", member)
					if err != nil {
						utilities.LogErr("Erro ao escrever em failed.txt: %s", err)
					}
					utilities.LogFailed("Token %v não conseguiu enviar DM para %v - Código de Erro: %v; Status: %v; Mensagem: %v", instances[i].CensorToken(), user, response.Code, respCode, response.Message)
					if cfg.OtherSettings.Logs {
						utilities.WriteLinesPath(logsFile, fmt.Sprintf("[%v][Sucesso:%v][Falha:%v] %v: falha ao enviar DM para %v, erro %v", time.Now().Format("15:04:05"), len(session), len(failed), instances[i].CensorToken(), user, response.Message))
					}
				}
				time.Sleep(time.Duration(cfg.DirectMessage.Delay) * time.Second)
				if cfg.SuspicionAvoidance.RandomIndividualDelay != 0 {
					time.Sleep(time.Duration(rand.Intn(cfg.SuspicionAvoidance.RandomIndividualDelay)) * time.Second)
				}
			}
		}(i)
	}
	wg.Wait()

	utilities.LogSuccess("Threads finalizaram! Escrevendo no arquivo...")
	ticker <- true
	elapsed := time.Since(start)
	utilities.LogSuccess("O envio de DMs demorou %.2f segundos. DMs enviadas com sucesso para %v IDs. Falha ao enviar para %v IDs. %v tokens estão disfuncionais & %v tokens estão funcionando", elapsed.Seconds(), len(session), len(failed), len(dead), len(instances)-len(dead))
	if cfg.OtherSettings.Logs {
		utilities.WriteLinesPath(logsFile, fmt.Sprintf("O envio de DMs demorou %.2f segundos. DMs enviadas com sucesso para %v IDs. Falha ao enviar para %v IDs. %v tokens estão disfuncionais & %v tokens estão funcionando", elapsed.Seconds(), len(session), len(failed), len(dead), len(instances)-len(dead)))
	}
	var left []string
	if cfg.DirectMessage.Remove {
		for i := 0; i < len(instances); i++ {
			if !utilities.Contains(dead, instances[i].Token) {
				if instances[i].Password == "" {
					left = append(left, instances[i].Token)
				} else {
					left = append(left, fmt.Sprintf(`%v:%v:%v`, instances[i].Email, instances[i].Password, instances[i].Token))
				}
			}
		}
		err := utilities.Truncate("input/tokens.txt", left)
		if err != nil {
			utilities.LogErr("Erro ao escrever em tokens.txt: %s", err)
		}
		utilities.LogSuccess("Arquivo tokens.txt atualizado.")
	}
	if cfg.DirectMessage.RemoveM {
		m := utilities.RemoveSubset(members, completed)
		err := utilities.Truncate("input/memberids.txt", m)
		if err != nil {
			utilities.LogErr("Erro ao escrever em memberids.txt: %s", err)
		}
		utilities.LogSuccess("Arquivo memberids.txt atualizado.")
	}
	if cfg.DirectMessage.Websocket {
		for i := 0; i < len(instances); i++ {
			if instances[i].Ws != nil {
				instances[i].Ws.Close()
			}
		}
	}
}

type jsonResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func LaunchSingleDM() {
	choice := utilities.UserInputInteger("Digite 0 para uma mensagem; Digite 1 para spam contínuo:")
	cfg, instances, err := instance.GetEverything()
	if err != nil {
		utilities.LogErr("Erro ao carregar configurações ou instâncias: %s", err)
		return
	}
	var msg instance.Message
	messagechoice := utilities.UserInputInteger("Digite 1 para usar a mensagem do arquivo, 2 para digitar no console: ")
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

	victim := utilities.UserInput("Garanta um servidor em comum e digite o ID da vítima: ")
	var wg sync.WaitGroup
	wg.Add(len(instances))
	if choice == 0 {
		for i := 0; i < len(instances); i++ {
			time.Sleep(time.Duration(cfg.DirectMessage.Offset) * time.Millisecond)
			go func(i int) {
				defer wg.Done()
				snowflake, err := instances[i].OpenChannel(victim)
				if err != nil {
					utilities.LogErr("Erro ao abrir o canal: %s", err)
				}
				respCode, body, err := instances[i].SendMessage(snowflake, victim)
				if err != nil {
					utilities.LogErr("Erro ao enviar mensagem: %s", err)
				}
				if respCode == 200 {
					utilities.LogSuccess("Token %v enviou DM para %v", instances[i].CensorToken(), victim)
				} else {
					utilities.LogFailed("Token %v falhou ao enviar DM para %v [%v]", instances[i].CensorToken(), victim, string(body))
				}
			}(i)
		}
		wg.Wait()
	}
	if choice == 1 {
		for i := 0; i < len(instances); i++ {
			time.Sleep(time.Duration(cfg.DirectMessage.Offset) * time.Millisecond)
			go func(i int) {
				defer wg.Done()

				var c int
				for {
					snowflake, err := instances[i].OpenChannel(victim)
					if err != nil {
						utilities.LogErr("Erro ao abrir o canal: %s", err)
					}
					respCode, _, err := instances[i].SendMessage(snowflake, victim)
					if err != nil {
						utilities.LogErr("Erro ao enviar mensagem: %s", err)
					}
					if respCode == 200 {
						utilities.LogSuccess("Token %v enviou DM para %v [%v]", instances[i].CensorToken(), victim, c)
					} else {
						utilities.LogFailed("Token %v falhou ao enviar DM para %v", instances[i].CensorToken(), victim)
					}
					c++
				}
			}(i)
		}
		wg.Wait()
	}
	utilities.LogSuccess("Todas as threads finalizaram.")
}