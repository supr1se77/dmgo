package discord

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/V4NSH4J/discord-mass-dm-GO/instance"
	"github.com/V4NSH4J/discord-mass-dm-GO/utilities"
	"github.com/zenthangplus/goccm"
)

func LaunchButtonClicker() {
	cfg, instances, err := instance.GetEverything()
	if err != nil {
		utilities.LogErr("Erro ao carregar instâncias ou configurações: %s", err)
		return
	}
	var tokenFile, successFile, failedFile string
	if cfg.OtherSettings.Logs {
		path := fmt.Sprintf(`logs/clicador_de_botao/LOGS-BOTAO-%s-%s`, time.Now().Format(`2006-01-02 15-04-05`), utilities.RandStringBytes(5))
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
		successFileX, err := os.Create(fmt.Sprintf(`%s/sucesso.txt`, path))
		if err != nil {
			utilities.LogErr("Erro ao criar arquivo de sucesso: %s", err)
			utilities.ExitSafely()
		}
		successFileX.Close()
		failedFileX, err := os.Create(fmt.Sprintf(`%s/falha.txt`, path))
		if err != nil {
			utilities.LogErr("Erro ao criar arquivo de falha: %s", err)
			utilities.ExitSafely()
		}
		failedFileX.Close()
		tokenFile, successFile, failedFile = tokenFileX.Name(), successFileX.Name(), failedFileX.Name()
		for i := 0; i < len(instances); i++ {
			instances[i].WriteInstanceToFile(tokenFile)
		}
	}
	token := utilities.UserInput("Digite um token que consegue ver a mensagem:")
	id := utilities.UserInput("Digite o ID da mensagem:")
	channel := utilities.UserInput("Digite o ID do canal:")
	server := utilities.UserInput("Digite o ID do servidor:")
	msg, err := instance.FindMessage(channel, id, token)
	if err != nil {
		utilities.LogErr("Erro ao encontrar a mensagem: %v", err)
		return
	}
	utilities.LogInfo("Mensagem encontrada!\n %s", msg)
	var Msg instance.Message
	err = json.Unmarshal([]byte(msg), &Msg)
	if err != nil {
		utilities.LogErr("Erro ao decodificar a mensagem: %v", err)
		return
	}
	if len(Msg.Components) == 0 {
		utilities.LogErr("A mensagem não tem componentes (botões, etc).")
		return
	}
	for i := 0; i < len(Msg.Components); i++ {
		fmt.Printf("%v) Linha %v\n", i, i)
	}
	row := utilities.UserInputInteger("Digite o número da linha:")
	for i := 0; i < len(Msg.Components[row].Buttons); i++ {
		if Msg.Components[row].Buttons[i].Label != "" {
			fmt.Printf("%v) Botão %v [%v]\n", i, i, Msg.Components[row].Buttons[i].Label)
		} else if Msg.Components[row].Buttons[i].Emoji.Name != "" {
			fmt.Printf("%v) Botão %v [%v]\n", i, i, Msg.Components[row].Buttons[i].Emoji)
		} else {
			fmt.Printf("%v) Botão %v [Nome ou Emoji não encontrado]\n", i, i)
		}
	}
	column := utilities.UserInputInteger("Selecione o Botão:")
	threads := utilities.UserInputInteger("Digite o número de threads:")
	if threads > len(instances) || threads == 0 {
		threads = len(instances)
	}
	c := goccm.New(threads)
	for i := 0; i < len(instances); i++ {
		c.Wait()
		go func(i int) {
			defer c.Done()
			err := instances[i].StartWS()
			if err != nil {
				utilities.LogFailed("Erro ao iniciar o websocket: %v", err)
			} else {
				utilities.LogSuccess("Websocket aberto para %s", instances[i].CensorToken())
			}
			respCode, err := instances[i].PressButton(row, column, server, Msg)
			if err != nil {
				utilities.LogFailed("Erro ao pressionar o botão: %v", err)
				if cfg.OtherSettings.Logs {
					utilities.WriteLinesPath(failedFile, instances[i].Token)
				}
				return
			}
			if respCode != 204 && respCode != 200 {
				utilities.LogFailed("Erro ao pressionar o botão, código: %v", respCode)
				if cfg.OtherSettings.Logs {
					utilities.WriteLinesPath(failedFile, instances[i].Token)
				}

				return
			}
			utilities.LogSuccess("Botão pressionado pela instância %v", instances[i].CensorToken())
			if cfg.OtherSettings.Logs {
				utilities.WriteLinesPath(successFile, instances[i].Token)
			}
			if instances[i].Ws != nil {
				if instances[i].Ws.Conn != nil {
					err = instances[i].Ws.Close()
					if err != nil {
						utilities.LogFailed("Erro ao fechar o websocket: %v", err)
					} else {
						utilities.LogSuccess("Websocket fechado para %v", instances[i].CensorToken())
					}
				}
			}
		}(i)
	}
	c.WaitAllDone()
}