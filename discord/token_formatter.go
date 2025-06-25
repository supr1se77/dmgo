// Copyright (C) 2021 github.com/V4NSH4J
//
// Esse código fonte foi liberado sob a licença GNU Affero General Public
// License v3.0. Uma cópia dessa licença tá aqui:
// https://www.gnu.org/licenses/agpl-3.0.en.html

package discord

import (
	"fmt"
	"os"
	"time"

	"github.com/V4NSH4J/discord-mass-dm-GO/instance"
	"github.com/V4NSH4J/discord-mass-dm-GO/utilities"
)

// Função que inicia o formatador de tokens
func LaunchTokenFormatter() {
	// Pega a configuração, as instâncias e erro (se tiver)
	cfg, instances, err := instance.GetEverything()
	if err != nil {
		// Loga o erro se deu merda ao pegar as infos necessárias
		utilities.LogErr("Erro pegando info necessária %v", err)
	}
	var tokenFile, changedFile string
	if cfg.OtherSettings.Logs {
		// Cria o caminho pra salvar os logs, com data e string randômica
		path := fmt.Sprintf(`logs/token_formatter/DMDGO-TF-%s-%s`, time.Now().Format(`2006-01-02 15-04-05`), utilities.RandStringBytes(5))
		// Cria o diretório, se der erro e não for porque já existe, loga e sai
		err := os.MkdirAll(path, 0755)
		if err != nil && !os.IsExist(err) {
			utilities.LogErr("Erro criando diretório de logs: %s", err)
			utilities.ExitSafely()
		}
		// Cria o arquivo token.txt dentro do diretório de logs
		tokenFileX, err := os.Create(fmt.Sprintf(`%s/token.txt`, path))
		if err != nil {
			utilities.LogErr("Erro criando arquivo de tokens: %s", err)
			utilities.ExitSafely()
		}
		tokenFileX.Close()
		// Cria o arquivo changed.txt para armazenar tokens alterados/sucesso
		ChangedFileX, err := os.Create(fmt.Sprintf(`%s/changed.txt`, path))
		if err != nil {
			utilities.LogErr("Erro criando arquivo de sucesso: %s", err)
			utilities.ExitSafely()
		}
		ChangedFileX.Close()
		// Salva os nomes dos arquivos para usar depois
		tokenFile, changedFile = tokenFileX.Name(), ChangedFileX.Name()
		// Pra cada instância, escreve no arquivo token.txt
		for i := 0; i < len(instances); i++ {
			instances[i].WriteInstanceToFile(tokenFile)
		}
	}
	// Cria slice pra guardar tokens puros
	var tokens []string

	// Loop nas instâncias pra limpar dados sensíveis e salvar no changed.txt se logs ativados
	for i := 0; i < len(instances); i++ {
		if cfg.OtherSettings.Logs {
			instances[i].Email = ""      // Zera o email pra não vazar nada
			instances[i].Password = ""   // Zera a senha pra mesma porra
			instances[i].WriteInstanceToFile(changedFile) // Salva no arquivo changed.txt
		}
		// Adiciona só o token na lista de tokens
		tokens = append(tokens, instances[i].Token)
	}
	// Salva essa lista de tokens no arquivo tokens.txt, truncando antes
	_ = utilities.TruncateLines("tokens.txt", tokens)
	// Loga que finalizou o formatador de tokens com sucesso
	utilities.LogSuccess("Formatador de tokens finalizado")
}
