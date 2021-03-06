package bot

import (
	"Votebot/votebot/cfg"
	"Votebot/votebot/db"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const helpText = `Votebot manages voting on what songs are available. Commands:
**!help** Prints this help text.
**!vote** {song number} Votes for a song by its number.
**!displayList** Prints the entire list of available songs.
**!search** {text} Search for a song with a sub-stirng of the title or artist.`
const sorryText = "Sorry, something went wrong there."

// Context holds the neccesary information to communicate with the server.
type Context struct {
	StartTime time.Time
	Session   *discordgo.Session
}

// Cxt holds the neccesary information to communicate with hte server.
var Cxt Context

// Init initiallizes the bot, by creating a session and registering the handler.
func Init() {

	Cxt = Context{}
	Cxt.StartTime = time.Now()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + cfg.Cfg.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	Cxt.Session = dg
	Cxt.Session.AddHandler(RouterHandler)

	// Open a websocket connection to Discord and begin listening.
	err = Cxt.Session.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
}

// validates the message and splits it into an array of substings split on space.
func parseMessage(m *discordgo.MessageCreate) ([]string, error) {

	if len(m.Content) < 5 {
		return nil, errors.New("message too short. No commands exist shorter than 5 chars")
	}
	substrings := strings.Split(m.Content, " ")

	return substrings, nil
}

func hasRole(member *discordgo.Member, role string) bool {

	for _, r := range member.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// RouterHandler is responsible for parsing the users command and rounding it to the correct function.
func RouterHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	command, err := parseMessage(m)
	if err != nil {
		fmt.Printf("Parsing a message threw an error: err %v", err)
		s.ChannelMessageSend(m.ChannelID, sorryText)
		return
	}

	// SetChannel is the only command that is not spesific to the slected channel. Special case.
	if m.ChannelID != cfg.Cfg.ChannelID {
		if strings.ToLower(command[0]) == "!setchannel" {
			err = SetChannel(command, s, m)
		}
		return
	}

	switch strings.ToLower(command[0]) {
	case "!hello":
		songs, err := db.GetAllSongs()
		if err != nil {
			fmt.Printf("Something went wrong with db: %v", err)
		}

		s.ChannelMessageSend(m.ChannelID, "World! Songs:"+strconv.Itoa(len(songs)))
		break
	case "!help":
		s.ChannelMessageSend(m.ChannelID, helpText)
		break

	case "!showCurrentSongs":

		break

	case "!showAllSongs":

		break

	case "!vote":

		break

	case "!search":

		break

	case "!openvotes":

		break

	case "!closevotes":

		break

	case "!setsonglimit":
		err = SetConfigInt(command, "Song limit", &cfg.Cfg.SongLimit, s, m)
		break

	case "!setvotecount":
		err = SetConfigInt(command, "Vote count", &cfg.Cfg.VotesPerUser, s, m)
		break

	case "!setpateronvotecount":
		err = SetConfigInt(command, "Patreon vote count", &cfg.Cfg.VotesPerPateron, s, m)
		break
	}

	if err != nil {
		fmt.Printf("A command threw an error: err %v", err)
		s.ChannelMessageSend(m.ChannelID, sorryText)
	}
}

// SetChannel changes the channelID the bot looks for messages in.
func SetChannel(command []string, s *discordgo.Session, m *discordgo.MessageCreate) error {

	if !hasRole(m.Member, cfg.Cfg.MasterRoleID) {
		return errors.New("user is missing required role to set channel")
	}

	channel, _ := Cxt.Session.Channel(m.ChannelID) // We are guarantied the channel exist.
	guildID := channel.GuildID
	if guildID == "" {
		fmt.Println("User tried to use setChannel in DMs.")
		return errors.New("can not set a text channel that is not within a guild")
	}
	cfg.Cfg.ChannelID = m.ChannelID
	cfg.WriteConfigFile()
	s.ChannelMessageSend(m.ChannelID, "New text channel set to "+channel.Name)
	fmt.Println("Text channel set to \"" + channel.Name + "\" in guild with ID: " + guildID)
	return nil
}

// SetConfigInt sets and integer in the config struct to a new a value and saves the config to disk.
func SetConfigInt(command []string, name string, value *int, s *discordgo.Session, m *discordgo.MessageCreate) error {

	if !hasRole(m.Member, cfg.Cfg.MasterRoleID) {
		return errors.New("user is missing required role to set vote count")
	}

	if len(command) < 2 {
		return errors.New("missing argument")
	}
	num, err := strconv.Atoi(command[1])
	if err != nil {
		return errors.New("unable to to parse input(" + name + ") to int")
	}

	*value = num
	cfg.WriteConfigFile()
	s.ChannelMessageSend(m.ChannelID, name+" set to "+command[1])
	fmt.Println(name + " set to " + command[1])
	return nil
}
