package main

import (
	_ "embed"
	"os"
	"time"

	"github.com/c0rydoras/slides/internal/model"
	"github.com/c0rydoras/slides/internal/navigation"
	"github.com/c0rydoras/slides/internal/preprocessor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	tocTitle       string
	tocDescription string
	enableHeadings bool

	cmd = &cobra.Command{
		Use:   "slides <file.md>",
		Short: "Terminal based presentation tool",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var fileName string

			if len(args) > 0 {
				fileName = args[0]
			}

			presentation := model.Model{
				Page:         0,
				Date:         time.Now().Format("2006-01-02"),
				FileName:     fileName,
				Search:       navigation.NewSearch(),
				Preprocessor: preprocessor.NewConfig().WithTOC(tocTitle, tocDescription).WithHeadings(),
			}
			err = presentation.Load()
			if err != nil {
				return err
			}

			p := tea.NewProgram(presentation, tea.WithAltScreen())
			_, err = p.Run()
			return err
		},
	}
)

func main() {
	cmd.Flags().BoolVarP(&enableHeadings, "headings", "a", false, "Enable automatic heading addition")

	cmd.Flags().StringVarP(&tocTitle, "toc", "t", "", "Enable table of contents generation with optional title (default: 'Table of Contents')")
	tocFlag := cmd.Flag("toc")
	tocFlag.NoOptDefVal = "Table of Contents"

	cmd.Flags().StringVarP(&tocDescription, "toc-description", "d", "", "Enable table of contents generation with optional description")
	tocDescFlag := cmd.Flag("toc-description")
	tocDescFlag.NoOptDefVal = "Table of Contents Description"

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
