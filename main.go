package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	integerOptionMinValue = 1.0

	commands = []*discordgo.ApplicationCommand{
		{
			Name: "ping",
			// All commands and options must have a description
			// Commands/options without description will fail the registration
			// of the command.
			Description: "Test responsiveness",
		},
		{
			Name: "teams",
			// All commands and options must have a description
			// Commands/options without description will fail the registration
			// of the command.
			Description: "Create or delete team pods",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "delete",
					Description: "Subcommands group",
					Options: []*discordgo.ApplicationCommandOption{
						// Also, subcommand groups aren't capable of
						// containing options, by the name of them, you can see
						// they can only contain subcommands
						{
							Name:        "by-role",
							Description: "Specific team to delete",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionRole,
									Name:        "team-role",
									Description: "Role for the team",
									Required:    true,
								},
							},
						},
						{
							Name:        "all",
							Description: "Delete all teams",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommandGroup,
				},
				{
					Name:        "create",
					Description: "Subcommands group",
					Options: []*discordgo.ApplicationCommandOption{
						// Also, subcommand groups aren't capable of
						// containing options, by the name of them, you can see
						// they can only contain subcommands
						{
							Name:        "by-name",
							Description: "Create a team with a specific name",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "team-name",
									Description: "Name for the team",
									Required:    true,
								},
							},
						},
						{
							Name:        "batch",
							Description: "Create a batch of teams",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionInteger,
									Name:        "team-count",
									Description: "Number of teams to create",
									Required:    true,
								},
							},
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommandGroup,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong",
				},
			})
		},
		"teams": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			channel, _ := s.Channel(i.ChannelID)
			guild, _ := s.Guild(channel.GuildID)
			options := i.ApplicationCommandData().Options
			title := "Bouncing back"
			content := "Starting"
			embed := makeEmbed(title, content)

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Embeds: embed},
			})
			// As you can see, names of subcommands (nested, top-level)
			// and subcommand groups are provided through the arguments.
			switch options[0].Name {
			case "create":
				options = options[0].Options
				if options[0] != nil {
					greenTeam := FindRoleByName(s, i, "Green Team")
					switch options[0].Name {
					case "by-name":
						options = options[0].Options
						optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
						for _, opt := range options {
							optionMap[opt.Name] = opt
						}
						// Initizlize embed
						teamName := optionMap["team-name"].StringValue()
						title = fmt.Sprintf("Creating team %s", teamName)
						status := "Initializing..."
						content = "**Single Job**\n"
						content += fmt.Sprintf("%s: %s\n", teamName, status)
						editEmbed(s, i, title, content)

						// Create role
						status = "Creating role..."
						content = "**Single Job**\n"
						content += fmt.Sprintf("%s: %s\n", teamName, status)
						editEmbed(s, i, title, content)
						var role, _ = s.GuildRoleCreate(guild.ID)
						s.GuildRoleEdit(
							guild.ID,
							role.ID,
							fmt.Sprintf("%s", teamName),
							3447003, // blue
							true,
							0,
							true,
						)
						// Create parent category
						status = "Creating channels..."
						content = "**Single Job**\n"
						content += fmt.Sprintf("%s: %s\n", teamName, status)
						editEmbed(s, i, title, content)
						var category, _ = s.GuildChannelCreateComplex(guild.ID, discordgo.GuildChannelCreateData{
							Name: fmt.Sprintf("%s", teamName),
							Type: 4,
							PermissionOverwrites: []*discordgo.PermissionOverwrite{
								{
									ID:   guild.ID,
									Type: 0,
									Deny: discordgo.PermissionViewChannel,
								},
								{
									ID:    role.ID,
									Type:  0,
									Allow: discordgo.PermissionViewChannel,
								},
								{
									ID:    greenTeam.ID,
									Type:  0,
									Allow: discordgo.PermissionViewChannel + discordgo.PermissionManageChannels,
								},
							},
						})
						// Create child channels
						s.GuildChannelCreateComplex(guild.ID, discordgo.GuildChannelCreateData{
							Name:     fmt.Sprintf("%s-text", teamName),
							Type:     0,
							ParentID: category.ID,
						})
						s.GuildChannelCreateComplex(guild.ID, discordgo.GuildChannelCreateData{
							Name:     fmt.Sprintf("%s-support", teamName),
							Type:     0,
							ParentID: category.ID,
						})
						s.GuildChannelCreateComplex(guild.ID, discordgo.GuildChannelCreateData{
							Name:     fmt.Sprintf("%s-voice", teamName),
							Type:     2,
							ParentID: category.ID,
						})
						status = "**Created successfully.**"
						content = "**Single Job**\n"
						content += fmt.Sprintf("%s: %s\n", teamName, status)
						editEmbed(s, i, title, content)

					case "batch":
						options = options[0].Options
						optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
						for _, opt := range options {
							optionMap[opt.Name] = opt
						}
						// Initialize embed
						teamCount := int(optionMap["team-count"].IntValue())
						title = fmt.Sprintf("Building %d teams", teamCount)
						statuses := make([]string, teamCount)
						content = "**Batch Create**\n"
						for j := 0; j < teamCount; j++ {
							statuses[j] = "Initializing..."
							content += fmt.Sprintf("Team %d: %s\n", j+1, statuses[j])
						}
						editEmbed(s, i, title, content)

						for j := 0; j < teamCount; j++ {
							// Create role
							statuses[j] = "Creating role..."
							batchUpdateEmbed(s, i, title, statuses)
							var role, _ = s.GuildRoleCreate(guild.ID)
							s.GuildRoleEdit(
								guild.ID,
								role.ID,
								fmt.Sprintf("Team %d", j+1),
								3447003, // blue
								true,
								0,
								true,
							)

							// Create parent category
							statuses[j] = "Creating channels..."
							batchUpdateEmbed(s, i, title, statuses)
							var category, _ = s.GuildChannelCreateComplex(guild.ID, discordgo.GuildChannelCreateData{
								Name: fmt.Sprintf("Team %d", j+1),
								Type: 4,
								PermissionOverwrites: []*discordgo.PermissionOverwrite{
									{
										ID:   guild.ID,
										Type: 0,
										Deny: discordgo.PermissionViewChannel,
									},
									{
										ID:    role.ID,
										Type:  0,
										Allow: discordgo.PermissionViewChannel,
									},
									{
										ID:    greenTeam.ID,
										Type:  0,
										Allow: discordgo.PermissionViewChannel + discordgo.PermissionManageChannels,
									},
								},
							})
							// Create child channels
							s.GuildChannelCreateComplex(guild.ID, discordgo.GuildChannelCreateData{
								Name:     fmt.Sprintf("team-%d-text", j+1),
								Type:     0,
								ParentID: category.ID,
							})
							s.GuildChannelCreateComplex(guild.ID, discordgo.GuildChannelCreateData{
								Name:     fmt.Sprintf("team-%d-support", j+1),
								Type:     0,
								ParentID: category.ID,
							})
							s.GuildChannelCreateComplex(guild.ID, discordgo.GuildChannelCreateData{
								Name:     fmt.Sprintf("team-%d-voice", j+1),
								Type:     2,
								ParentID: category.ID,
							})
							statuses[j] = "**Done.**"
							batchUpdateEmbed(s, i, title, statuses)
						}
					default:
						title = "Error"
						content = "Oops, something went wrong.\n" +
							"Hol' up, you aren't supposed to see this message."
						editEmbed(s, i, title, content)
					}
				}

			case "delete":
				options = options[0].Options
				var role *discordgo.Role
				if options[0] != nil {
					switch options[0].Name {
					case "by-role":
						options = options[0].Options
						optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
						for _, opt := range options {
							optionMap[opt.Name] = opt
						}
						// Initialize embed
						role = optionMap["team-role"].RoleValue(s, guild.ID)
						title = fmt.Sprintf("Deleting pod for %s", role.Name)
						content = "**Single Team Remove**\n"
						status := "Finding..."
						content += fmt.Sprintf("%s: %s", role.Name, status)
						editEmbed(s, i, title, content)

						// Find parent based on name
						roles, _ := s.GuildRoles(guild.ID)
						channels, _ := s.GuildChannels(guild.ID)
						parent := FindChannelByName(s, i, role.Name)

						// Remove channels
						content = "**Single Team Remove**\n"
						status = "Deleting role..."
						content += fmt.Sprintf("%s: %s", role.Name, status)
						editEmbed(s, i, title, content)
						for _, gchannel := range channels {
							if gchannel.ParentID == parent.ID {
								s.ChannelDelete(gchannel.ID)
							}
						}
						s.ChannelDelete(parent.ID)
						// Remove role
						content = "**Single Team Remove**\n"
						status = "Deleting role..."
						content += fmt.Sprintf("%s: %s", role.Name, status)
						editEmbed(s, i, title, content)
						for _, grole := range roles {
							if grole.ID == role.ID {
								s.GuildRoleDelete(guild.ID, grole.ID)
							}
						}

						// Clean up
						content = "**Single Team Remove**\n"
						status = "**Removed successfully.**"
						content += fmt.Sprintf("%s: %s", role.Name, status)
						editEmbed(s, i, title, content)
					case "all":
						groles, _ := s.GuildRoles(guild.ID)
						channels, _ := s.GuildChannels(guild.ID)
						roles := make([]*discordgo.Role, 0)
						for _, role := range groles {
							// Exclude protected names
							var match, _ = regexp.MatchString("(.?everyone|Green Team|Red Team|RvBMo|Public)", role.Name)
							if !match {
								roles = append(roles, role)
							}
						}

						// Initialize embed
						statuses := make([]string, len(roles))
						title = fmt.Sprintf("Deleting all (%d) pods", len(roles))
						content = "**Batch Remove**\n"
						for index, _ := range statuses {
							statuses[index] = "Finding..."
							content += fmt.Sprintf("%s: %s", roles[index].Name, statuses[index])
						}
						editEmbed(s, i, title, content)

						for index, teamRole := range roles {
							// Find parent based on name
							parent := FindChannelByName(s, i, teamRole.Name)

							// Remove channels
							statuses[index] = "Deleting role..."
							deleteAllUpdateEmbed(s, i, title, statuses, roles)
							for _, gchannel := range channels {
								if gchannel.ParentID == parent.ID {
									s.ChannelDelete(gchannel.ID)
								}
							}
							s.ChannelDelete(parent.ID)

							// Remove role
							statuses[index] = "Deleting role..."
							deleteAllUpdateEmbed(s, i, title, statuses, roles)

							for _, grole := range roles {
								if grole.ID == teamRole.ID {
									s.GuildRoleDelete(guild.ID, grole.ID)
								}
							}

							// Clean up
							statuses[index] = "**Removed successfully.**"
							deleteAllUpdateEmbed(s, i, title, statuses, roles)
						}
					default:
						title = "Error"
						content = "Oops, something went wrong.\n" +
							"Hol' up, you aren't supposed to see this message."
						editEmbed(s, i, title, content)
					}
				}

			default:
				title = "Error"
				content = "Oops, something went wrong.\n" +
					"Hol' up, you aren't supposed to see this message."
				editEmbed(s, i, title, content)

			}
		},
	}
)

func batchUpdateEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, title string, statuses []string) {
	content := "**Batch Create**\n"
	for team := 0; team < len(statuses); team++ {
		content += fmt.Sprintf("Team %d: %s\n", team+1, statuses[team])
	}
	editEmbed(s, i, title, content)
}

func deleteAllUpdateEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, title string, statuses []string, roles []*discordgo.Role) {
	content := "**Batch Remove**\n"
	for index, _ := range statuses {
		content += fmt.Sprintf("%s: %s\n", roles[index].Name, statuses[index])
	}
	editEmbed(s, i, title, content)
}
func editEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, title string, content string) {
	log.Printf("%s", content)
	embed := makeEmbed(title, content)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: embed})
}

func makeEmbed(title string, content string) []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{
		{
			Type:        "rich",
			Title:       fmt.Sprintf("RvBMo • %s", title),
			Description: content,
			Color:       16755520,
		},
	}
}

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func FindChannelByName(s *discordgo.Session, i *discordgo.InteractionCreate, channelName string) *discordgo.Channel {
	var channel *discordgo.Channel
	c, _ := s.Channel(i.ChannelID)
	guild, _ := s.Guild(c.GuildID)
	channels, _ := s.GuildChannels(guild.ID)
	for j := 0; j < len(channels); j++ {
		log.Printf("%s", (channels[j].Name))
		if channels[j].Name == channelName {
			channel = channels[j]
		}
	}
	return channel
}

func FindRoleByName(s *discordgo.Session, i *discordgo.InteractionCreate, rolename string) *discordgo.Role {
	var role *discordgo.Role
	channel, _ := s.Channel(i.ChannelID)
	guild, _ := s.Guild(channel.GuildID)
	for j := 0; j < len(guild.Roles); j++ {
		if guild.Roles[j].Name == rolename {
			role = guild.Roles[j]
		}
	}
	return role
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
		log.Printf("Added \"%s\"", v.Name)
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		log.Println("Removing commands...")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
