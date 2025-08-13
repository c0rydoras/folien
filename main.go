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
	cmd.Flags().BoolVarP(&enableHeadings, "headings", "a", false, "Enable automatic heading addition")

	cmd.Flags().StringVarP(&tocTitle, "toc", "t", "", "Enable table of contents generation with optional title (default: 'Table of Contents')")
	tocFlag := cmd.Flag("toc")
	tocFlag.NoOptDefVal = "Table of Contents"

	cmd.Flags().StringVarP(&tocDescription, "toc-description", "d", "", "Enable table of contents generation with optional description")
	tocDescFlag := cmd.Flag("toc-description")
	tocDescFlag.NoOptDefVal = "Table of Contents Description"
}

var cmd = &cobra.Command{
	Use:               "folien <file.md>",
	Short:             "Terminal based presentation tool",
	Args:              cobra.ExactArgs(1),
	RunE:              root,
	ValidArgsFunction: cobra.FixedCompletions(nil, cobra.ShellCompDirectiveDefault|cobra.ShellCompDirectiveFilterFileExt),
}

func main() {
	if err := fang.Execute(context.Background(), cmd); err != nil {
		os.Exit(1)
	}
}

func root(cmd *cobra.Command, args []string) error {
	var err error
	fileName := args[0]

	preprocessorConfig := preprocessor.NewConfig().WithTOC(tocTitle, tocDescription)
	if enableHeadings {
		preprocessorConfig = preprocessorConfig.WithHeadings()
	}

	presentation := model.Model{
		Page:         0,
		Date:         time.Now().Format("2006-01-02"),
		FileName:     fileName,
		Search:       navigation.NewSearch(),
		Preprocessor: preprocessorConfig,
	}
	err = presentation.Load()
	if err != nil {
		return err
	}

	p := tea.NewProgram(presentation, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
