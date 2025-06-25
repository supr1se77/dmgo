package discord

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/V4NSH4J/discord-mass-dm-GO/instance"
	"github.com/V4NSH4J/discord-mass-dm-GO/utilities"
)

func LaunchScraperMenu() {
	cfg, _, err := instance.GetEverything()
	if err != nil {
		utilities.LogErr("Erro ao carregar as informações necessárias: %v", err)
		utilities.ExitSafely()
	}
	utilities.PrintMenu([]string{"Extração Online (Opcode 14)", "Extração por Reações (API REST)", "Extração Offline (Opcode 8)"})
	options := utilities.UserInputInteger("Selecione uma opção: ")
	if options == 1 {
		token := utilities.UserInput("Digite o token: ")
		serverid := utilities.UserInput("Digite o ID do servidor: ")
		channelid := utilities.UserInput("Digite o ID do canal: ")
		var botsFile, avatarFile, nameFile, path, rolePath, scrapedFile, userDataFile string
		if cfg.OtherSettings.Logs {
			path = fmt.Sprintf(`logs/extracao_online/LOGS-ONLINE-%s-%s-%s-%s`, serverid, channelid, time.Now().Format(`2006-01-02 15-04-05`), utilities.RandStringBytes(5))
			rolePath = fmt.Sprintf(`%s/cargos`, path)
			err := os.MkdirAll(path, 0755)
			if err != nil && !os.IsExist(err) {
				utilities.LogErr("Erro ao criar o diretório de logs: %s", err)
				utilities.ExitSafely()
			}
			err = os.MkdirAll(rolePath, 0755)
			if err != nil && !os.IsExist(err) {
				utilities.LogErr("Erro ao criar o diretório de cargos: %s", err)
				utilities.ExitSafely()
			}
			botsFileX, err := os.Create(fmt.Sprintf(`%s/bots.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de bots: %s", err)
				utilities.ExitSafely()
			}
			botsFileX.Close()
			AvatarFileX, err := os.Create(fmt.Sprintf(`%s/avatares.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de avatares: %s", err)
				utilities.ExitSafely()
			}
			AvatarFileX.Close()
			NameFileX, err := os.Create(fmt.Sprintf(`%s/nomes.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de nomes: %s", err)
				utilities.ExitSafely()
			}
			NameFileX.Close()
			ScrapedFileX, err := os.Create(fmt.Sprintf(`%s/extraidos.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de extraídos: %s", err)
				utilities.ExitSafely()
			}
			ScrapedFileX.Close()
			UserDataFileX, err := os.Create(fmt.Sprintf(`%s/dados_usuarios.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de dados de usuários: %s", err)
				utilities.ExitSafely()
			}
			UserDataFileX.Close()
			botsFile, avatarFile, nameFile, scrapedFile, userDataFile = botsFileX.Name(), AvatarFileX.Name(), NameFileX.Name(), ScrapedFileX.Name(), UserDataFileX.Name()
		}
		Is := instance.Instance{Token: token}
		title := make(chan bool)
		go func() {
		Out:
			for {
				select {
				case <-title:
					break Out
				default:
					if Is.Ws != nil {
						if Is.Ws.Conn != nil {
							cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf(`SUPR1SE [%v Extraídos]`, len(Is.Ws.Members)))
							_ = cmd.Run()
						}
					}
				}
			}
		}()
		t := 0
		for {
			if t >= 5 {
				utilities.LogErr("Não foi possível conectar ao websocket após várias tentativas.")
				break
			}
			err := Is.StartWS()
			if err != nil {
				utilities.LogFailed("Erro ao abrir o websocket: %v", err)
			} else {
				break
			}
			t++
		}

		utilities.LogSuccess("Websocket aberto para %v", Is.CensorToken())
		i := 0
		for {
			err := instance.Scrape(Is.Ws, serverid, channelid, i)
			if err != nil {
				utilities.LogErr("Erro ao extrair: %v", err)
			}
			utilities.LogSuccess("Token %v | Contagem de Extração: %v", Is.CensorToken(), len(Is.Ws.Members))
			if Is.Ws.Complete {
				break
			}
			i++
			time.Sleep(time.Duration(cfg.ScraperSettings.SleepSc) * time.Millisecond)
		}
		if Is.Ws != nil {
			Is.Ws.Close()
		}
		utilities.LogSuccess("Extração finalizada. %v membros extraídos.", len(Is.Ws.Members))
		if cfg.OtherSettings.Logs {
			for i := 0; i < len(Is.Ws.Members); i++ {
				if Is.Ws.Members[i].User.Bot {
					utilities.WriteLinesPath(botsFile, fmt.Sprintf("%v %v %v", Is.Ws.Members[i].User.ID, Is.Ws.Members[i].User.Username, Is.Ws.Members[i].User.Discriminator))
				}
				if Is.Ws.Members[i].User.Avatar != "" {
					utilities.WriteLinesPath(avatarFile, fmt.Sprintf("%v:%v", Is.Ws.Members[i].User.ID, Is.Ws.Members[i].User.Avatar))
				}
				if Is.Ws.Members[i].User.Username != "" {
					utilities.WriteLinesPath(nameFile, fmt.Sprintf("%v", Is.Ws.Members[i].User.Username))
				}
				for x := 0; x < len(Is.Ws.Members[i].Roles); x++ {
					utilities.WriteRoleFile(Is.Ws.Members[i].User.ID, rolePath, Is.Ws.Members[i].Roles[x])
				}
				if Is.Ws.Members[i].User.Discriminator != "" && Is.Ws.Members[i].User.Username != "" {
					utilities.WriteLinesPath(userDataFile, fmt.Sprintf("%v#%v", Is.Ws.Members[i].User.Username, Is.Ws.Members[i].User.Discriminator))
				}
				utilities.WriteLinesPath(scrapedFile, Is.Ws.Members[i].User.ID)
			}
		}
		var memberids []string
		for _, member := range Is.Ws.Members {
			memberids = append(memberids, member.User.ID)
		}
		clean := utilities.RemoveDuplicateStr(memberids)
		utilities.LogSuccess("Duplicados removidos. %v membros extraídos.", len(clean))
		write := utilities.UserInput("Deseja escrever em memberids.txt? (s/n)")
		title <- true
		if write == "s" || write == "S" {
			for k := 0; k < len(clean); k++ {
				err := utilities.WriteLines("memberids.txt", clean[k])
				if err != nil {
					utilities.LogErr("Erro ao escrever no arquivo: %v", err)
				}
			}
			utilities.LogSuccess("Escreveu %v membros em memberids.txt", len(clean))
		}

	}
	if options == 2 {
		token := utilities.UserInput("Digite o token: ")
		channelid := utilities.UserInput("Digite o ID do canal: ")
		messageid := utilities.UserInput("Digite o ID da mensagem: ")
		utilities.PrintMenu([]string{"Obter Emoji da Mensagem", "Digitar Emoji Manualmente"})
		option := utilities.UserInputInteger("Selecione uma opção: ")
		var send string
		if option == 2 {
			send = utilities.UserInput("Digite o emoji [Formato nomeEmoji ou nomeEmoji:IDemoji para emojis nitro]: ")
		} else {
			msg, err := instance.GetRxn(channelid, messageid, token)
			if err != nil {
				utilities.LogErr("Erro ao obter a mensagem: %v", err)
			}
			var selection []string
			for i := 0; i < len(msg.Reactions); i++ {
				selection = append(selection, fmt.Sprintf("Emoji: %v | Contagem: %v", msg.Reactions[i].Emojis.Name, msg.Reactions[i].Count))
			}
			utilities.PrintMenu2(selection)
			index := utilities.UserInputInteger("Selecione uma opção: ")
			if msg.Reactions[index].Emojis.ID == "" {
				send = msg.Reactions[index].Emojis.Name

			} else if msg.Reactions[index].Emojis.ID != "" {
				send = msg.Reactions[index].Emojis.Name + ":" + msg.Reactions[index].Emojis.ID
			}
		}
		var allUIDS []string
		var m string
		title := make(chan bool)
		go func() {
		Out:
			for {
				select {
				case <-title:
					break Out
				default:
					cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf(`SUPR1SE [%v Extraídos]`, len(allUIDS)))
					_ = cmd.Run()
				}
			}
		}()
		for {
			if len(allUIDS) == 0 {
				m = ""
			} else {
				m = allUIDS[len(allUIDS)-1]
			}
			rxn, err := instance.GetReactions(channelid, messageid, token, send, m)
			if err != nil {
				utilities.LogErr("Erro ao obter as reações: %v", err)
				continue
			}
			if len(rxn) == 0 {
				break
			}
			utilities.LogInfo("Extraídos %v membros", len(rxn))
			allUIDS = append(allUIDS, rxn...)
		}
		utilities.LogInfo("Extração finalizada. %v linhas extraídas - Removendo duplicados", len(allUIDS))
		clean := utilities.RemoveDuplicateStr(allUIDS)
		path := fmt.Sprintf(`logs/extracao_reacoes/LOGS-REACAO-%s-%s-%s-%s`, channelid, messageid, time.Now().Format(`2006-01-02 15-04-05`), utilities.RandStringBytes(5))
		err := os.MkdirAll(path, 0755)
		if err != nil && !os.IsExist(err) {
			utilities.LogErr("Erro ao criar o diretório de logs: %s", err)
			utilities.ExitSafely()
		}
		scrapedFileX, err := os.Create(fmt.Sprintf(`%s/extraidos.txt`, path))
		if err != nil {
			utilities.LogErr("Erro ao criar arquivo de extraídos: %s", err)
			utilities.ExitSafely()
		}
		defer scrapedFileX.Close()
		scrapedFile := scrapedFileX.Name()
		for i := 0; i < len(clean); i++ {
			utilities.WriteLinesPath(scrapedFile, clean[i])
		}
		write := utilities.UserInput("Deseja escrever em memberids.txt? (s/n)")
		if write == "s" || write == "S" {
			for k := 0; k < len(clean); k++ {
				err := utilities.WriteLines("memberids.txt", clean[k])
				if err != nil {
					utilities.LogErr("Erro ao escrever no arquivo: %v", err)
				}
			}
			utilities.LogSuccess("Escreveu %v membros em memberids.txt", len(clean))
		}
		title <- true
		utilities.LogSuccess("Finalizado.")
	}
	if options == 3 {
		cfg, instances, err := instance.GetEverything()
		if err != nil {
			utilities.LogErr("Erro ao carregar instâncias: %v", err)
			return
		}
		var scraped []string
		var queriesCompleted []string
		title := make(chan bool)
		go func() {
		Out:
			for {
				select {
				case <-title:
					break Out
				default:
					cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf(`SUPR1SE [%v Extraídos | %v Buscas Concluídas]`, len(scraped), len(queriesCompleted)))
					_ = cmd.Run()
				}
			}
		}()
		numTokens := utilities.UserInputInteger("Quantos tokens deseja usar? Você tem %v: ", len(instances))
		quit := make(chan bool)
		var allQueries []string
		var chars string
		rawChars := " !\"#$%&'()*+,-./0123456789:;<=>?@[]^_`abcdefghijklmnopqrstuvwxyz{|}~" + cfg.ScraperSettings.ExtendedChars
		for i := 0; i < len(rawChars); i++ {
			if !strings.Contains(rawChars[0:i], string(rawChars[i])) {
				chars += string(rawChars[i])
			}
		}
		queriesLeft := make(chan string)
		for i := 0; i < len(chars); i++ {
			go func(i int) {
				queriesLeft <- string(chars[i])
			}(i)
		}
		if numTokens > len(instances) {
			utilities.LogWarn("Você só tem %v tokens em tokens.txt. Usando o número máximo de tokens possível.", len(instances))
		} else if numTokens <= 0 {
			utilities.LogErr("Você deve usar pelo menos 1 token.")
			utilities.ExitSafely()
		} else if numTokens <= len(instances) {
			utilities.LogInfo("Você tem %v tokens em tokens.txt. Usando %v tokens.", len(instances), numTokens)
			instances = instances[:numTokens]
		} else {
			utilities.LogErr("Entrada inválida.")
		}
		serverid := utilities.UserInput("Digite o ID do servidor: ")
		var tokenFile, botsFile, avatarFile, nameFile, path, rolePath, scrapedFile, userDataFile string
		if cfg.OtherSettings.Logs {
			path = fmt.Sprintf(`logs/extracao_offline/LOGS-OFFLINE-%s-%s-%s`, serverid, time.Now().Format(`2006-01-02 15-04-05`), utilities.RandStringBytes(5))
			rolePath = fmt.Sprintf(`%s/cargos`, path)
			err := os.MkdirAll(path, 0755)
			if err != nil && !os.IsExist(err) {
				utilities.LogErr("Erro ao criar o diretório de logs: %s", err)
				utilities.ExitSafely()
			}
			err = os.MkdirAll(rolePath, 0755)
			if err != nil && !os.IsExist(err) {
				utilities.LogErr("Erro ao criar o diretório de cargos: %s", err)
				utilities.ExitSafely()
			}
			tokenFileX, err := os.Create(fmt.Sprintf(`%s/token.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de token: %s", err)
				utilities.ExitSafely()
			}
			tokenFileX.Close()
			botsFileX, err := os.Create(fmt.Sprintf(`%s/bots.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de bots: %s", err)
				utilities.ExitSafely()
			}
			botsFileX.Close()
			AvatarFileX, err := os.Create(fmt.Sprintf(`%s/avatares.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de avatares: %s", err)
				utilities.ExitSafely()
			}
			AvatarFileX.Close()
			NameFileX, err := os.Create(fmt.Sprintf(`%s/nomes.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de nomes: %s", err)
				utilities.ExitSafely()
			}
			NameFileX.Close()
			ScrapedFileX, err := os.Create(fmt.Sprintf(`%s/extraidos.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de extraídos: %s", err)
				utilities.ExitSafely()
			}
			defer ScrapedFileX.Close()
			UserDataFileX, err := os.Create(fmt.Sprintf(`%s/dados_usuarios.txt`, path))
			if err != nil {
				utilities.LogErr("Erro ao criar arquivo de dados de usuários: %s", err)
				utilities.ExitSafely()
			}
			defer UserDataFileX.Close()
			tokenFile, botsFile, avatarFile, nameFile, scrapedFile, userDataFile = tokenFileX.Name(), botsFileX.Name(), AvatarFileX.Name(), NameFileX.Name(), ScrapedFileX.Name(), UserDataFileX.Name()
			for i := 0; i < len(instances); i++ {
				instances[i].WriteInstanceToFile(tokenFile)
			}
		}
		utilities.LogInfo("Pressione ENTER para INICIAR e PARAR a extração")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		var namesScraped []string
		var avatarsScraped []string
		for i := 0; i < len(instances); i++ {
			go func(i int) {
				instances[i].ScrapeCount = 0
				for {
					if instances[i].ScrapeCount%5 == 0 || instances[i].LastCount%100 == 0 {
						if instances[i].Ws != nil {
							instances[i].Ws.Close()
						}
						time.Sleep(2 * time.Second)
						err := instances[i].StartWS()
						if err != nil {
							fmt.Println(err)
							continue
						}
						time.Sleep(2 * time.Second)
					}
					instances[i].ScrapeCount++

					select {
					case <-quit:
						if instances[i].Ws != nil {
							instances[i].Ws.Close()
						}
						return
					default:
						query := <-queriesLeft
						allQueries = append(allQueries, query)
						if instances[i].Ws == nil {
							continue
						}
						if instances[i].Ws.Conn == nil {
							continue
						}
						err := instance.ScrapeOffline(instances[i].Ws, serverid, query)
						if err != nil {
							utilities.LogErr("%v: Erro ao extrair: %v", instances[i].CensorToken(), err)
							go func() {
								queriesLeft <- query
							}()
							continue
						}

						memInfo := <-instances[i].Ws.OfflineScrape
						queriesCompleted = append(queriesCompleted, query)
						var MemberInfo instance.Event
						err = json.Unmarshal(memInfo, &MemberInfo)
						if err != nil {
							utilities.LogErr("Erro ao decodificar: %v", err)
							queriesLeft <- query
							continue
						}

						if len(MemberInfo.Data.Members) == 0 {
							instances[i].LastCount = -1
							continue
						}
						instances[i].LastCount = len(MemberInfo.Data.Members)
						for _, member := range MemberInfo.Data.Members {
							if !utilities.Contains(scraped, member.User.ID) {
								scraped = append(scraped, member.User.ID)
							}
						}
						utilities.LogSuccess("Token %v | Busca %v | Extraídos %v [+%v]", instances[i].CensorToken(), query, len(scraped), len(MemberInfo.Data.Members))

						for i := 0; i < len(MemberInfo.Data.Members); i++ {
							id := MemberInfo.Data.Members[i].User.ID
							err := utilities.WriteLines("memberids.txt", id)
							if err != nil {
								utilities.LogErr("Erro ao escrever no arquivo: %v", err)
								continue
							}
							if cfg.OtherSettings.Logs {
								utilities.WriteLinesPath(scrapedFile, id)
								if MemberInfo.Data.Members[i].User.Bot {
									utilities.WriteLinesPath(botsFile, fmt.Sprintf("%v %v %v", id, MemberInfo.Data.Members[i].User.Username, MemberInfo.Data.Members[i].User.Discriminator))
								}
								if MemberInfo.Data.Members[i].User.Avatar != "" {
									utilities.WriteLinesPath(avatarFile, fmt.Sprintf("%v:%v", id, MemberInfo.Data.Members[i].User.Avatar))
								}
								if MemberInfo.Data.Members[i].User.Username != "" {
									utilities.WriteLinesPath(nameFile, fmt.Sprintf("%v", MemberInfo.Data.Members[i].User.Username))
								}
								for x := 0; x < len(MemberInfo.Data.Members[i].Roles); x++ {
									utilities.WriteRoleFile(id, rolePath, MemberInfo.Data.Members[i].Roles[x])
								}
								if MemberInfo.Data.Members[i].User.Username != "" && MemberInfo.Data.Members[i].User.Discriminator != "" {
									utilities.WriteLinesPath(userDataFile, fmt.Sprintf("%v#%v", MemberInfo.Data.Members[i].User.Username, MemberInfo.Data.Members[i].User.Discriminator))
								}
							}

							if cfg.ScraperSettings.ScrapeUsernames {
								nom := MemberInfo.Data.Members[i].User.Username
								if !utilities.Contains(namesScraped, nom) {
									err := utilities.WriteLines("names.txt", nom)
									if err != nil {
										utilities.LogErr("Erro ao escrever no arquivo: %v", err)
										continue
									}
								}
							}
							if cfg.ScraperSettings.ScrapeAvatars {
								av := MemberInfo.Data.Members[i].User.Avatar
								if !utilities.Contains(avatarsScraped, av) {
									err := utilities.ProcessAvatar(av, id)
									if err != nil {
										utilities.LogErr("Erro ao processar avatar: %v", err)
										continue
									}
								}
							}
						}
						if len(MemberInfo.Data.Members) < 100 {
							time.Sleep(time.Duration(cfg.ScraperSettings.SleepSc) * time.Millisecond)
							continue
						}
						lastName := MemberInfo.Data.Members[len(MemberInfo.Data.Members)-1].User.Username

						nextQueries := instance.FindNextQueries(query, lastName, queriesCompleted, chars)
						for i := 0; i < len(nextQueries); i++ {
							go func(i int) {
								queriesLeft <- nextQueries[i]
							}(i)
						}
					}
				}
			}(i)
		}
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		utilities.LogInfo("Parando todas as instâncias...")
		title <- true
		for i := 0; i < len(instances); i++ {
			go func() {
				quit <- true
			}()
		}
		utilities.LogInfo("Extração Completa. %v membros extraídos.", len(scraped))
		choice := utilities.UserInput("Deseja escrever no arquivo novamente? (s/n) [Isso removerá os IDs pré-existentes em memberids.txt]")
		if choice == "s" || choice == "S" {
			clean := utilities.RemoveDuplicateStr(scraped)
			err := utilities.TruncateLines("memberids.txt", clean)
			if err != nil {
				utilities.LogErr("Erro ao limpar o arquivo: %v", err)
			}
		}
	}
}