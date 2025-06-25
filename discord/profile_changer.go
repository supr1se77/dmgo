// Copyright (C) 2021 github.com/V4NSH4J
//
// Este código fonte foi liberado sob a Licença Pública Geral Affero GNU v3.0.
// Uma cópia desta licença está disponível em
// https://www.gnu.org/licenses/agpl-3.0.en.html

package discord

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/V4NSH4J/discord-mass-dm-GO/instance"
	"github.com/V4NSH4J/discord-mass-dm-GO/utilities"
	"github.com/zenthangplus/goccm"
)

// IniciarTrocadorDeNome inicia o processo de alteração de nome de usuário para as contas.
func LaunchNameChanger() {
	_, instancias, err := instance.GetEverything()
	if err != nil {
		utilities.LogErr("Erro ao pegar a configuração ou as instâncias: %v", err)
	}
	var contagemTotal, contagemSucesso, contagemFalha int
	titulo := make(chan bool)
	go func() {
	Fora:
		for {
			select {
			case <-titulo:
				break Fora
			default:
				// Atualiza o título da janela do console com o progresso.
				cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf(`DMDGO [%v Sucesso, %v Falha, %v Não processado]`, contagemSucesso, contagemFalha, contagemTotal-contagemSucesso-contagemFalha))
				_ = cmd.Run()
			}

		}
	}()
	for i := 0; i < len(instancias); i++ {
		if instancias[i].Password == "" {
			utilities.LogWarn("O Token %v não tem senha definida. O trocador de nome precisa do token no formato email:senha:token", instancias[i].CensorToken())
			continue
		}
	}
	utilities.LogWarn("Os nomes de usuário são trocados aleatoriamente a partir do arquivo.")
	usuarios, err := utilities.ReadLines("names.txt")
	if err != nil {
		utilities.LogErr("Erro ao ler o arquivo names.txt: %v", err)
		return
	}
	if len(usuarios) == 0 {
		utilities.LogErr("O arquivo names.txt está vazio.")
		return
	}
	threads := utilities.UserInputInteger("Digite o número de threads (0 para o máximo):")
	if threads > len(instancias) || threads == 0 {
		threads = len(instancias)
	}
	contagemTotal = len(instancias)
	c := goccm.New(threads)
	for i := 0; i < len(instancias); i++ {
		c.Wait()
		go func(i int) {
			err := instancias[i].StartWS()
			if err != nil {
				utilities.LogErr("Token %v - Erro ao abrir o websocket: %v", instancias[i].CensorToken(), err)
			} else {
				utilities.LogSuccess("Token %v - Websocket aberto", instancias[i].CensorToken())
			}
			r, err := instancias[i].NameChanger(usuarios[rand.Intn(len(usuarios))])
			if err != nil {
				utilities.LogErr("Token %v - Erro ao trocar o nome: %v", instancias[i].CensorToken(), err)
				contagemFalha++
				return
			}
			body, err := utilities.ReadBody(r)
			if err != nil {
				utilities.LogErr("Token %v - Erro ao ler o corpo da resposta: %v", instancias[i].CensorToken(), err)
				contagemFalha++
				return
			}
			if r.StatusCode == 200 || r.StatusCode == 204 {
				utilities.LogSuccess("Token %v - Nome trocado com sucesso", instancias[i].CensorToken())
				contagemSucesso++
			} else {
				utilities.LogFailed("Token %v - Erro ao trocar o nome: %v %v", instancias[i].CensorToken(), r.Status, string(body))
				contagemFalha++
			}
			if instancias[i].Ws != nil {
				if instancias[i].Ws.Conn != nil {
					err = instancias[i].Ws.Close()
					if err != nil {
						utilities.LogFailed("Token %v - Erro ao fechar o websocket: %v", instancias[i].CensorToken(), err)
					} else {
						utilities.LogSuccess("Token %v - Websocket fechado", instancias[i].CensorToken())
					}
					c.Done()
				}
			}
		}(i)
	}
	c.WaitAllDone()
	titulo <- true
	utilities.LogSuccess("Trocador de nomes finalizado.")

}

// IniciarTrocadorDeAvatar inicia o processo de alteração de avatar para as contas.
func LaunchAvatarChanger() {
	_, instancias, err := instance.GetEverything()
	if err != nil {
		utilities.LogErr("Erro ao pegar a configuração ou as instâncias: %v", err)
	}
	var contagemTotal, contagemSucesso, contagemFalha int
	titulo := make(chan bool)
	go func() {
	Fora:
		for {
			select {
			case <-titulo:
				break Fora
			default:
				cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf(`DMDGO [%v Sucesso, %v Falha, %v Não processado]`, contagemSucesso, contagemFalha, contagemTotal-contagemSucesso-contagemFalha))
				_ = cmd.Run()
			}
		}
	}()
	utilities.LogWarn("AVISO: Apenas PNG e JPEG/JPG são suportados. As fotos de perfil são trocadas aleatoriamente da pasta. Use o formato PNG para resultados mais rápidos.")
	utilities.LogInfo("Carregando avatares...")
	ex, err := os.Executable()
	if err != nil {
		utilities.LogErr("Erro ao obter o caminho do executável: %v", err)
		utilities.ExitSafely()
	}
	ex = filepath.ToSlash(ex)
	caminho := path.Join(path.Dir(ex) + "/input/pfps")

	imagens, err := instance.GetFiles(caminho)
	if err != nil {
		// Se der erro, cria o diretório para o usuário.
		utilities.LogErr("Erro ao obter arquivos de %v: %v", caminho, err)
		utilities.LogInfo("Criando a pasta 'input/pfps' para você...")
		err = os.MkdirAll(caminho, os.ModePerm)
		if err != nil {
			utilities.LogErr("Não foi possível criar a pasta 'input/pfps': %v", err)
			utilities.ExitSafely()
		}
		utilities.LogSuccess("Pasta 'input/pfps' criada. Por favor, adicione as imagens lá e rode a ferramenta novamente.")
		utilities.ExitSafely()
	}

	if len(imagens) == 0 {
		utilities.LogErr("Nenhuma imagem encontrada na pasta 'input/pfps'.")
		utilities.ExitSafely()
	}

	utilities.LogInfo("%v arquivos carregados", len(imagens))
	var avatares []string

	for i := 0; i < len(imagens); i++ {
		av, err := instance.EncodeImg(imagens[i])
		if err != nil {
			utilities.LogErr("Erro ao codificar a imagem %v: %v", imagens[i], err)
			continue
		}
		avatares = append(avatares, av)
	}
	utilities.LogInfo("%v avatares carregados", len(avatares))
	threads := utilities.UserInputInteger("Digite o número de threads (0 para o máximo):")
	if threads > len(instancias) || threads == 0 {
		threads = len(instancias)
	}
	contagemTotal = len(instancias)
	c := goccm.New(threads)
	for i := 0; i < len(instancias); i++ {
		c.Wait()
		go func(i int) {
			err := instancias[i].StartWS()
			if err != nil {
				utilities.LogFailed("Token %v - Erro ao abrir o websocket", instancias[i].CensorToken())
			} else {
				utilities.LogSuccess("Websocket aberto %v", instancias[i].CensorToken())
			}
			r, err := instancias[i].AvatarChanger(avatares[rand.Intn(len(avatares))])
			if err != nil {
				utilities.LogFailed("Token %v - Erro ao trocar o avatar: %v", instancias[i].CensorToken(), err)
				contagemFalha++
			} else {
				if r.StatusCode == 204 || r.StatusCode == 200 {
					utilities.LogSuccess("Token %v - Avatar trocado com sucesso", instancias[i].CensorToken())
					contagemSucesso++
				} else {
					utilities.LogFailed("Token %v - Erro ao trocar o avatar: %v", instancias[i].CensorToken(), r.StatusCode)
					contagemFalha++
				}
			}
			if instancias[i].Ws != nil {
				if instancias[i].Ws.Conn != nil {
					err = instancias[i].Ws.Close()
					if err != nil {
						utilities.LogFailed("Token %v - Erro ao fechar o websocket: %v", instancias[i].CensorToken(), err)
					} else {
						utilities.LogSuccess("Token %v - Websocket fechado", instancias[i].CensorToken())
					}
					c.Done()
				}
			}

		}(i)
	}
	c.WaitAllDone()
	titulo <- true
	utilities.LogSuccess("Trocador de avatares finalizado.")
}

// IniciarTrocadorDeBio inicia o processo de alteração de biografia para as contas.
func LaunchBioChanger() {
	bios, err := utilities.ReadLines("bios.txt")
	if err != nil {
		utilities.LogErr("Erro ao ler o arquivo bios.txt: %v", err)
		utilities.ExitSafely()
	}
	_, instancias, err := instance.GetEverything()
	if err != nil {
		utilities.LogErr("Erro ao pegar a configuração ou as instâncias: %v", err)
		utilities.ExitSafely()
	}
	var contagemTotal, contagemSucesso, contagemFalha int
	titulo := make(chan bool)
	go func() {
	Fora:
		for {
			select {
			case <-titulo:
				break Fora
			default:
				cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf(`DMDGO [%v Sucesso, %v Falha, %v Não processado]`, contagemSucesso, contagemFalha, contagemTotal-contagemSucesso-contagemFalha))
				_ = cmd.Run()
			}

		}
	}()
	bios = instance.ValidateBios(bios)
	utilities.LogInfo("%v bios carregadas, %v instâncias", len(bios), len(instancias))
	threads := utilities.UserInputInteger("Digite o número de threads (0 para o máximo):")
	if threads > len(instancias) || threads == 0 {
		threads = len(instancias)
	}
	contagemTotal = len(instancias)
	c := goccm.New(threads)
	for i := 0; i < len(instancias); i++ {
		c.Wait()
		go func(i int) {
			err := instancias[i].StartWS()
			if err != nil {
				utilities.LogFailed("Token %v - Erro ao abrir o websocket", instancias[i].CensorToken())
			} else {
				utilities.LogSuccess("Token %v - Websocket aberto", instancias[i].CensorToken())
			}
			err = instancias[i].BioChanger(bios)
			if err != nil {
				utilities.LogFailed("%v - Erro ao trocar a bio: %v", instancias[i].CensorToken(), err)
				contagemFalha++
			} else {
				utilities.LogSuccess("%v - Bio trocada com sucesso", instancias[i].CensorToken())
				contagemSucesso++
			}
			if instancias[i].Ws != nil {
				if instancias[i].Ws.Conn != nil {
					err = instancias[i].Ws.Close()
					if err != nil {
						utilities.LogFailed("Token %v - Erro ao fechar o websocket: %v", instancias[i].CensorToken(), err)
					} else {
						utilities.LogSuccess("Token %v - Websocket fechado", instancias[i].CensorToken())
					}
					c.Done()
				}
			}
		}(i)
	}
	titulo <- true
	c.WaitAllDone()
	utilities.LogSuccess("Trocador de bios finalizado.")
}

// IniciarTrocadorDeHypeSquad inicia o processo de alteração de HypeSquad para as contas.
func LaunchHypeSquadChanger() {
	_, instancias, err := instance.GetEverything()
	if err != nil {
		utilities.LogErr("Erro ao pegar a configuração ou as instâncias: %v", err)
		utilities.ExitSafely()
	}
	var contagemTotal, contagemSucesso, contagemFalha int
	titulo := make(chan bool)
	go func() {
	Fora:
		for {
			select {
			case <-titulo:
				break Fora
			default:
				cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf(`DMDGO [%v Sucesso, %v Falha, %v Não processado]`, contagemSucesso, contagemFalha, contagemTotal-contagemSucesso-contagemFalha))
				_ = cmd.Run()
			}

		}
	}()
	threads := utilities.UserInputInteger("Digite o número de threads (0 para o máximo):")
	if threads > len(instancias) || threads == 0 {
		threads = len(instancias)
	}
	contagemTotal = len(instancias)
	c := goccm.New(threads)
	for i := 0; i < len(instancias); i++ {
		c.Wait()
		go func(i int) {
			err := instancias[i].RandomHypeSquadChanger()
			if err != nil {
				utilities.LogFailed("Token %v - Erro ao trocar a HypeSquad: %v", instancias[i].CensorToken(), err)
				contagemFalha++
			} else {
				utilities.LogSuccess("Token %v - HypeSquad trocada com sucesso", instancias[i].CensorToken())
				contagemSucesso++
			}
			c.Done()
		}(i)
	}
	titulo <- true
	c.WaitAllDone()
	utilities.LogSuccess("Trocador de HypeSquad finalizado.")
}

// IniciarTrocadorDeToken inicia o processo de alteração de senha e token para as contas.
func LaunchTokenChanger() {

	_, instancias, err := instance.GetEverything()
	if err != nil {
		utilities.LogErr("Erro ao pegar a configuração ou as instâncias: %v", err)
	}
	var contagemTotal, contagemSucesso, contagemFalha int
	titulo := make(chan bool)
	go func() {
	Fora:
		for {
			select {
			case <-titulo:
				break Fora
			default:
				cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf(`DMDGO [%v Trocado, %v Falhou, %v Não Processado]`, contagemSucesso, contagemFalha, contagemTotal-contagemSucesso-contagemFalha))
				_ = cmd.Run()
			}

		}
	}()
	for i := 0; i < len(instancias); i++ {
		if instancias[i].Password == "" {
			utilities.LogWarn("%v - Nenhuma senha definida. Pode estar formatado incorretamente. O único formato suportado é email:senha:token", instancias[i].CensorToken())
			continue
		}
	}
	modo := utilities.UserInputInteger("Digite 0 para trocar as senhas aleatoriamente e 1 para trocar por uma senha fixa:")

	if modo != 0 && modo != 1 {
		utilities.LogErr("Modo inválido")
		utilities.ExitSafely()
	}
	var senha string
	if modo == 1 {
		senha = utilities.UserInput("Digite a senha para a qual os tokens serão alterados:")
	}
	threads := utilities.UserInputInteger("Digite o número de threads (0 para o máximo):")
	if threads > len(instancias) || threads == 0 {
		threads = len(instancias)
	}
	contagemTotal = len(instancias)
	c := goccm.New(threads)
	for i := 0; i < len(instancias); i++ {
		c.Wait()
		go func(i int) {
			if senha == "" {
				senha = utilities.RandStringBytes(12)
			}
			novoToken, err := instancias[i].ChangeToken(senha)
			if err != nil {
				utilities.LogFailed("Token %v - Erro ao trocar o token: %v", instancias[i].CensorToken(), err)
				contagemFalha++
				err := utilities.WriteLine("input/changed_tokens.txt", fmt.Sprintf(`%s:%s:%s`, instancias[i].Email, instancias[i].Password, instancias[i].Token))
				if err != nil {
					utilities.LogErr("Erro ao escrever no arquivo: %v", err)
				}
			} else {
				utilities.LogSuccess("%v - Token trocado com sucesso", instancias[i].CensorToken())
				contagemSucesso++
				err := utilities.WriteLine("input/changed_tokens.txt", fmt.Sprintf(`%s:%s:%s`, instancias[i].Email, senha, novoToken))
				if err != nil {
					utilities.LogErr("Erro ao escrever no arquivo: %v", err)
				}
			}
			c.Done()
		}(i)
	}
	c.WaitAllDone()
	titulo <- true
	utilities.LogSuccess("Trocador de tokens finalizado.")

}

// IniciarTrocadorDeApelidoServidor inicia o processo de alteração de apelido em um servidor específico.
func LaunchServerNicknameChanger() {
	_, instancias, err := instance.GetEverything()
	if err != nil {
		utilities.LogErr("Erro ao pegar a configuração ou as instâncias: %v", err)
	}
	var contagemTotal, contagemSucesso, contagemFalha int
	titulo := make(chan bool)
	go func() {
	Fora:
		for {
			select {
			case <-titulo:
				break Fora
			default:
				cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf(`DMDGO [%v Sucesso, %v Falha, %v Não processado]`, contagemSucesso, contagemFalha, contagemTotal-contagemSucesso-contagemFalha))
				_ = cmd.Run()
			}

		}
	}()
	utilities.LogWarn("AVISO: Os apelidos são trocados aleatoriamente a partir do arquivo.")
	apelidos, err := utilities.ReadLines("nicknames.txt")
	if err != nil {
		utilities.LogErr("Erro ao ler o arquivo nicknames.txt: %v", err)
		utilities.ExitSafely()
	}

	idServidor := utilities.UserInput("Digite o ID do servidor:")

	threads := utilities.UserInputInteger("Digite o número de threads (0 para o máximo):")
	if threads > len(instancias) || threads == 0 {
		threads = len(instancias)
	}
	contagemTotal = len(instancias)
	c := goccm.New(threads)
	for i := 0; i < len(instancias); i++ {
		c.Wait()
		go func(i int) {
			r, err := instancias[i].NickNameChanger(apelidos[rand.Intn(len(apelidos))], idServidor)
			if err != nil {
				utilities.LogFailed("Token %v - Erro ao trocar o apelido: %v", instancias[i].CensorToken(), err)
				contagemFalha++
				return
			}
			body, err := utilities.ReadBody(r)
			if err != nil {
				fmt.Println(err)
			}
			if r.StatusCode == 200 || r.StatusCode == 204 {
				utilities.LogSuccess("Token %v - Apelido trocado com sucesso", instancias[i].CensorToken())
				contagemSucesso++
			} else {
				utilities.LogFailed("Token %v - Erro ao trocar o apelido: %v %v", instancias[i].CensorToken(), r.Status, string(body))
				contagemFalha++
			}
			c.Done()
		}(i)
	}
	c.WaitAllDone()
	titulo <- true
	utilities.LogSuccess("Tudo Pronto.")

}

// IniciarSpammerDePedidosDeAmizade inicia o processo de envio em massa de pedidos de amizade.
func LaunchFriendRequestSpammer() {
	_, instancias, err := instance.GetEverything()
	if err != nil {
		utilities.LogErr("Erro ao pegar a configuração ou as instâncias: %v", err)
		return
	}
	var contagemTotal, contagemSucesso, contagemFalha int
	titulo := make(chan bool)
	go func() {
	Fora:
		for {
			select {
			case <-titulo:
				break Fora
			default:
				cmd := exec.Command("cmd", "/C", "title", fmt.Sprintf(`DMDGO [%v Sucesso, %v Falha, %v Não processado]`, contagemSucesso, contagemFalha, contagemTotal-contagemSucesso-contagemFalha))
				_ = cmd.Run()
			}

		}
	}()
	threads := utilities.UserInputInteger("Digite o número de threads (0 para o máximo):")
	if threads > len(instancias) || threads == 0 {
		threads = len(instancias)
	}
	nomeUsuario := utilities.UserInput("Digite o nome de usuário para spammar (Apenas o nome, sem a tag):")
	tag := utilities.UserInputInteger("Digite a tag para spammar (Apenas os números, sem o nome):")
	contagemTotal = len(instancias)
	c := goccm.New(threads)
	for i := 0; i < len(instancias); i++ {
		c.Wait()
		go func(i int) {
			defer c.Done()
			r, err := instancias[i].Friend(nomeUsuario, tag)
			if err != nil {
				utilities.LogFailed("Token %v - Erro ao enviar pedido de amizade: %v", instancias[i].CensorToken(), err)
				contagemFalha++
				return
			}
			body, err := utilities.ReadBody(*r)
			if err != nil {
				utilities.LogErr("Erro ao ler o corpo da resposta: %v", err)
				contagemFalha++
				return
			}
			if r.StatusCode == 200 || r.StatusCode == 204 {
				utilities.LogSuccess("Token %v - Pedido de amizade enviado com sucesso", instancias[i].CensorToken())
				contagemSucesso++
			} else {
				utilities.LogFailed("Token %v - Erro ao enviar pedido de amizade: %v %v", instancias[i].CensorToken(), r.Status, string(body))
				contagemFalha++
			}
		}(i)
	}
	c.WaitAllDone()
	titulo <- true
	utilities.LogSuccess("Tudo Pronto.")
}