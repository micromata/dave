package subcmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"syscall"
)

var passwdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "Generates a BCrypt hash of a given input string",
	Run: func(cmd *cobra.Command, args []string) {
		pw1 := readPassword()
		pw2 := readPassword()

		pw1Str := string(pw1)
		pw2Str := string(pw2)

		if pw1Str != pw2Str {
			fmt.Println("Passwords doesn't match.")
			os.Exit(1)
		}

		fmt.Printf("Hashed Password: %s\n", GenHash(pw1))
	},
}

func GenHash(password []byte) string {
	pw, err := bcrypt.GenerateFromPassword(password, 10)
	if err != nil {
		log.Fatal(err)
	}

	return string(pw)
}

func readPassword() []byte {
	fmt.Print("Enter password: ")
	pw, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("An error occurred reading the password: %s\n", err)
		os.Exit(1)
	}

	fmt.Println()
	return pw
}

func init() {
	RootCmd.AddCommand(passwdCmd)
}
