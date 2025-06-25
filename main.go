package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/V4NSH4J/discord-mass-dm-GO/discord"
	"github.com/V4NSH4J/discord-mass-dm-GO/utilities"
	"github.com/gookit/color"
)

func main() {
	version := "1.11.2"
	rand.Seed(time.Now().UTC().UnixNano())
	// Cor do logo alterada para Magenta (Roxo)
	color.Magenta.Printf(logo + " v" + version + "\n")
	// Cor dos créditos alterada para combinar com o tema
	color.LightMagenta.Printf("\t\t\t\t   SUPR1SE\n\n")
	utilities.VersionCheck(version)
	Options()
}

// Menu de Opções
func Options() {
	utilities.PrintMenu([]string{
		"Entrar em Servidores", "DM em Massa", "DM Único", "Adicionar Reação",
		"Converter Email:Senha:Token", "Verificador de Token", "Sair de Servidores",
		"Deixar Token Online", "Menu de Extração", "Mudar Nome", "Mudar Avatar",
		"Verificar Servidores do Token", "Mudar Biografia", "DM por Reação",
		"Mudar Hypesquad", "Mudar Senha do Token", "Criador de Embed",
		"Logar no Token", "Nuker de Token", "Pressionador de Botão",
		"Mudar Apelido no Servidor", "Spammer de Pedidos de Amizade",
		"DM em Massa para Amigos", "Novo DM em Massa [ALPHA]",
		"Créditos & Ajuda", "Sair",
	})
	choice := utilities.UserInputInteger("\t\tDigite sua escolha: ")
	switch choice {
	default:
		color.Red.Printf("\t\tOpção inválida!\n")
		Options()
	case 1:
		color.White.Printf("\t\tIniciando 'Entrar em Servidores'...\n")
		discord.LaunchinviteJoiner()
	case 2:
		color.White.Printf("\t\tIniciando 'DM em Massa'...\n")
		discord.LaunchMassDM()
	case 3:
		color.White.Printf("\t\tIniciando 'DM Único'...\n")
		discord.LaunchSingleDM()
	case 4:
		color.White.Printf("\t\tIniciando 'Adicionar Reação'...\n")
		discord.LaunchReactionAdder()
	case 5:
		color.White.Printf("\t\tIniciando 'Converter Email:Senha:Token'...\n")
		discord.LaunchTokenFormatter()
	case 6:
		color.White.Printf("\t\tIniciando 'Verificador de Token'...\n")
		discord.LaunchTokenChecker()
	case 7:
		color.White.Printf("\t\tIniciando 'Sair de Servidores'...\n")
		discord.LaunchGuildLeaver()
	case 8:
		color.White.Printf("\t\tIniciando 'Deixar Token Online'...\n")
		discord.LaunchTokenOnliner()
	case 9:
		color.White.Printf("\t\tAbrindo 'Menu de Extração'...\n")
		discord.LaunchScraperMenu()
	case 10:
		color.White.Printf("\t\tIniciando 'Mudar Nome'...\n")
		discord.LaunchNameChanger()
	case 11:
		color.White.Printf("\t\tIniciando 'Mudar Avatar'...\n")
		discord.LaunchAvatarChanger()
	case 12:
		color.White.Printf("\t\tIniciando 'Verificar Servidores do Token'...\n")
		discord.LaunchServerChecker()
	case 13:
		color.White.Printf("\t\tIniciando 'Mudar Biografia'...\n")
		discord.LaunchBioChanger()
	case 14:
		color.White.Printf("\t\tIniciando 'DM por Reação'...\n")
		discord.LaunchDMReact()
	case 15:
		color.White.Printf("\t\tIniciando 'Mudar Hypesquad'...\n")
		discord.LaunchHypeSquadChanger()
	case 16:
		color.White.Printf("\t\tIniciando 'Mudar Senha do Token'...\n")
		discord.LaunchTokenChanger()
	case 17:
		color.White.Printf("\t\tIniciando 'Criador de Embed'...\n")
		discord.LanuchEmbed()
	case 18:
		color.White.Printf("\t\tIniciando 'Logar no Token'...\n")
		discord.LaunchTokenLogin()
	case 19:
		color.White.Printf("\t\tIniciando 'Nuker de Token'...\n")
		discord.LaunchTokenNuker()
	case 20:
		color.White.Printf("\t\tIniciando 'Pressionador de Botão'...\n")
		discord.LaunchButtonClicker()
	case 21:
		color.White.Printf("\t\tIniciando 'Mudar Apelido no Servidor'...\n")
		discord.LaunchServerNicknameChanger()
	case 22:
		color.White.Printf("\t\tIniciando 'Spammer de Pedidos de Amizade'...\n")
		discord.LaunchFriendRequestSpammer()
	case 23:
		color.White.Printf("\t\tIniciando 'DM em Massa para Amigos'...\n")
		discord.LaunchFriendSpammer()
	case 24:
		color.White.Printf("\t\tIniciando 'Novo DM em Massa [ALPHA]'...\n")
		discord.LaunchAntiAntiRaidMode()
	case 25:
		color.Magenta.Printf("\t\tSUPR1SE\n")
	case 26:
		os.Exit(0)
	}
	time.Sleep(1 * time.Second)
	Options()

}

// Logo foi modificado com espaços para um efeito de centralização
const logo = `
		██████╗ ███╗   ███╗██████╗  ██████╗  ██████╗
		██╔══██╗████╗ ████║██╔══██╗██╔════╝ ██╔═══██╗
		██║  ██║██╔████╔██║██║  ██║██║  ███╗██║   ██║
		██║  ██║██║╚██╔╝██║██║  ██║██║   ██║██║   ██║
		██████╔╝██║ ╚═╝ ██║██████╔╝╚██████╔╝╚██████╔╝
		╚═════╝ ╚═╝     ╚═╝╚═════╝  ╚═════╝  ╚═════╝
			      FERRAMENTA DO SUPR1SE`