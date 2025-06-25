// Copyright (C) 2021 github.com/V4NSH4J
//
// Esse código foi liberado sob a licença GNU Affero General Public
// License v3.0. Dá uma conferida aqui pra ler a bagaça:
// https://www.gnu.org/licenses/agpl-3.0.en.html

package discord

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/V4NSH4J/discord-mass-dm-GO/instance"
	"github.com/V4NSH4J/discord-mass-dm-GO/utilities"
)

// Função pra rodar o foda-se e checar se o token tá no servidor
func LaunchServerChecker() {
	// Pega as configs, instâncias e erro, se pá deu ruim
	cfg, instances, err := instance.GetEverything()
	if err != nil {
		// Se ferrar ao pegar info, joga o erro e sai fora
		utilities.LogErr("Erro pegando info necessária %v", err)
		return
	}
	var tokenFile, presentFile, notPresentFile string
	if cfg.OtherSettings.Logs {
		// Cria o caminho dos logs com data e string aleatória pra não dar merda
		path := fmt.Sprintf(`logs/server_checker/DMDGO-SC-%s-%s`, time.Now().Format(`2006-01-02 15-04-05`), utilities.RandStringBytes(5))
		// Cria o diretório dos logs, se der erro e não for pq já existe, explode o bicho
		err := os.MkdirAll(path, 0755)
		if err != nil && !os.IsExist(err) {
			utilities.LogErr("Erro criando diretório de logs: %s", err)
			utilities.ExitSafely()
		}
		// Cria arquivo token.txt pra salvar os tokens
		tokenFileX, err := os.Create(fmt.Sprintf(`%s/token.txt`, path))
		if err != nil {
			utilities.LogErr("Erro criando arquivo de tokens: %s", err)
			utilities.ExitSafely()
		}
		tokenFileX.Close()
		// Cria arquivo present.txt pra salvar os que tão no servidor
		presentFileX, err := os.Create(fmt.Sprintf(`%s/present.txt`, path))
		if err != nil {
			utilities.LogErr("Erro criando arquivo present.txt: %s", err)
			utilities.ExitSafely()
		}
		presentFileX.Close()
		// Cria arquivo not_present.txt pra salvar os fracassados que não tão no servidor
		notPresentFileX, err := os.Create(fmt.Sprintf(`%s/not_present.txt`, path))
		if err != nil {
			utilities.LogErr("Erro criando arquivo not_present.txt: %s", err)
			utilities.ExitSafely()
		}
		notPresentFileX.Close()
		// Salva os nomes dos arquivos pra usar depois
		tokenFile, presentFile, notPresentFile = tokenFileX.Name(), presentFileX.Name(), notPresentFileX.Name()
		// Escreve cada instância no arquivo token.txt
		for i := 0; i < len(instances); i++ {
			instances[i].WriteInstanceToFile(tokenFile)
		}
	}
	var serverid string
	var inServer []string
	// Canal pra manipular o título do CMD
	title := make(chan bool)
	go func() {
	Out:
		for {
			select {
			case <-title:
				break Out
			default:
				// Muda o título do prompt com quantos tokens tão no servidor, pra dar visual foda
				cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf(`DMDGO [%v Present in Server]`, len(inServer)))
				_ = cmd.Run()
			}

		}
	}()
	// Pede pro usuário digitar o ID do servidor que quer checar
	serverid = utilities.UserInput("Digite o ID do servidor: ")
	var wg sync.WaitGroup
	// Cria um waitgroup pra esperar todo mundo terminar o trampo
	wg.Add(len(instances))
	for i := 0; i < len(instances); i++ {
		go func(i int) {
			defer wg.Done()
			// Checa se o token da instância tá no servidor
			r, err := instances[i].ServerCheck(serverid)
			if err != nil {
				// Se deu erro, loga o erro, marca como não presente se os logs tão ativados
				utilities.LogErr("%v Erro checando servidor: %v", instances[i].CensorToken(), err)
				if cfg.OtherSettings.Logs {
					instances[i].WriteInstanceToFile(notPresentFile)
				}
			} else {
				// Se retornou 200 ou 204, o filho da puta tá no servidor
				if r == 200 || r == 204 {
					utilities.LogSuccess("%v tá no servidor %v", instances[i].CensorToken(), serverid)
					inServer = append(inServer, instances[i].Token)
					if cfg.OtherSettings.Logs {
						instances[i].WriteInstanceToFile(presentFile)
					}
				} else if r == 429 {
					// Rate limited, ou seja, fodeu, tá bloqueado temporariamente
					utilities.LogFailed("%v tá rate limited", instances[i].CensorToken())
					if cfg.OtherSettings.Logs {
						instances[i].WriteInstanceToFile(notPresentFile)
					}
				} else if r == 400 {
					// Pedido mal feito, servidor inválido
					utilities.LogFailed("Pedido mal feito - ID do servidor inválido")
					if cfg.OtherSettings.Logs {
						instances[i].WriteInstanceToFile(notPresentFile)
					}
				} else {
					// Qualquer outro código é que o token não tá no servidor
					utilities.LogFailed("%v não tá no servidor [%v]", instances[i].CensorToken(), serverid, r)
					if cfg.OtherSettings.Logs {
						instances[i].WriteInstanceToFile(notPresentFile)
					}
				}
			}
		}(i)
	}
	// Espera todos terminarem o trampo
	wg.Wait()
	// Manda parar de atualizar o título
	title <- true
	// Pergunta se o usuário quer salvar os resultados
	save := utilities.UserInput("Quer salvar os resultados? (y/n)")
	if save == "y" || save == "Y" {
		// Salva os tokens que tão no servidor no arquivo tokens.txt truncado
		err := utilities.TruncateLines("tokens.txt", inServer)
		if err != nil {
			utilities.LogErr("Erro truncando arquivo: %v", err)
		} else {
			utilities.LogSuccess("Arquivo truncado com sucesso")
		}
	}
}
