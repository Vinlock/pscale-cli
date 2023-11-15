package password

import (
	"fmt"

	"github.com/planetscale/cli/internal/cmdutil"
	"github.com/planetscale/cli/internal/printer"
	ps "github.com/planetscale/planetscale-go/planetscale"

	"github.com/spf13/cobra"
)

func CreateCmd(ch *cmdutil.Helper) *cobra.Command {
	var flags struct {
		role string
	}

	createReq := &ps.DatabaseBranchPasswordRequest{}
	cmd := &cobra.Command{
		Use:     "create <database> <branch> <name>",
		Short:   "Create password to access a branch's data",
		Args:    cmdutil.RequiredArgs("database", "branch", "name"),
		Aliases: []string{"p"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			database := args[0]
			branch := args[1]
			name := args[2]

			if flags.role != "" {
				_, err := cmdutil.RoleFromString(flags.role)
				if err != nil {
					return err
				}
			}

			createReq.Database = database
			createReq.Branch = branch
			createReq.Organization = ch.Config.Organization
			createReq.Name = name
			createReq.Role = flags.role

			client, err := ch.Client()
			if err != nil {
				return err
			}

			end := ch.Printer.PrintProgress(fmt.Sprintf("Creating password of %s/%s...", printer.BoldBlue(database), printer.BoldBlue(branch)))
			defer end()

			pass, err := client.Passwords.Create(ctx, createReq)
			if err != nil {
				switch cmdutil.ErrCode(err) {
				case ps.ErrNotFound:
					return fmt.Errorf("branch %s does not exist in database %s (organization: %s)",
						printer.BoldBlue(branch), printer.BoldBlue(database), printer.BoldBlue(ch.Config.Organization))
				default:
					return cmdutil.HandleError(err)
				}
			}

			end()
			if ch.Printer.Format() == printer.Human {
				saveWarning := printer.BoldRed("Please save the values below as they will not be shown again")
				ch.Printer.Printf("Password %s was successfully created in %s/%s.\n%s\n\n",
					printer.BoldBlue(pass.Name), printer.BoldBlue(database), printer.BoldBlue(branch), saveWarning)
			}

			return ch.Printer.PrintResource(toPasswordWithPlainText(pass))
		},
	}
	cmd.PersistentFlags().StringVar(&flags.role, "role",
		"admin", "Role defines the access level, allowed values are : reader, writer, readwriter, admin. By default it is admin.")
	cmd.PersistentFlags().IntVar(&createReq.TTL, "ttl", 0, "TTL defines the time to live for the password in seconds. By default it is 0 which means it will never expire.")
	return cmd
}
