package main

import (
	"context"
	_ "embed"
	"os"
	"time"

	"github.com/c0rydoras/folien/internal/model"
	"github.com/c0rydoras/folien/internal/navigation"
	"github.com/c0rydoras/folien/internal/preprocessor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

var (
	tocTitle       string
	tocDescription string
	enableHeadings bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&enableHeadings, "headings", "a", false, "Enable automatic heading addition")

	rootCmd.PersistentFlags().StringVarP(&tocTitle, "toc", "t", "", "Enable table of contents generation with optional title (default: 'Table of Contents')")
	tocFlag := rootCmd.Flag("toc")
	tocFlag.NoOptDefVal = "Table of Contents"

	rootCmd.PersistentFlags().StringVarP(&tocDescription, "toc-description", "d", "", "Enable table of contents generation with optional description")
	tocDescFlag := rootCmd.Flag("toc-description")
	tocDescFlag.NoOptDefVal = "Table of Contents Description"

	rootCmd.AddCommand(serveCmd)
}

var rootCmd = &cobra.Command{
	Use:               "folien <file.md>",
	Short:             "Terminal based presentation tool",
	Args:              cobra.RangeArgs(0, 1),
	RunE:              root,
	ValidArgsFunction: cobra.FixedCompletions(nil, cobra.ShellCompDirectiveDefault|cobra.ShellCompDirectiveFilterFileExt),
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

func main() {
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

func root(cmd *cobra.Command, args []string) error {
	var err error
	if len(args) != 1 {
		return cmd.Help()
	}
	fileName := args[0]

	preprocessorConfig := preprocessor.NewConfig().WithTOC(tocTitle, tocDescription)
	if enableHeadings {
		preprocessorConfig = preprocessorConfig.WithHeadings()
	}

	presentation := model.Model{
		Page:               0,
		Date:               time.Now().Format("2006-01-02"),
		FileName:           fileName,
		Search:             navigation.NewSearch(),
		Preprocessor:       preprocessorConfig,
		HideInternalErrors: model.AllButLast,
	}
	err = presentation.Load()
	if err != nil {
		return err
	}

	p := tea.NewProgram(
		presentation,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
