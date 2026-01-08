package scanner

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"pdf_generator/internal/models"
)

var orderRegex = regexp.MustCompile(`^(\d+(\.\d+)*)`)

func ScanDirectory(root string) ([]models.Chapter, error) {
	var chapters []models.Chapter

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			name := info.Name()
			order := "999" // Default high order
			title := strings.TrimSuffix(name, ".md")

			match := orderRegex.FindString(name)
			if match != "" {
				order = match
				title = strings.TrimSpace(strings.TrimPrefix(title, match))
				title = strings.TrimPrefix(title, "_")
				title = strings.TrimPrefix(title, "-")
			}

			chapters = append(chapters, models.Chapter{
				Title:   title,
				Content: string(content),
				Path:    path,
				Order:   order,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort chapters by order string
	sort.Slice(chapters, func(i, j int) bool {
		return compareOrders(chapters[i].Order, chapters[j].Order)
	})

	return chapters, nil
}

func compareOrders(a, b string) bool {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	for i := 0; i < len(partsA) && i < len(partsB); i++ {
		if partsA[i] != partsB[i] {
			// Compare numerically if possible
			var valA, valB int
			_, errA := fmtSscanf(partsA[i], "%d", &valA)
			_, errB := fmtSscanf(partsB[i], "%d", &valB)

			if errA == nil && errB == nil {
				if valA != valB {
					return valA < valB
				}
			}
			return partsA[i] < partsB[i]
		}
	}
	return len(partsA) < len(partsB)
}

// Simple helper to avoid fmt import just for Sscanf if not needed, 
// but actually fmt is standard. I'll use it.
func fmtSscanf(str, format string, a ...interface{}) (int, error) {
	return fmt.Sscanf(str, format, a...)
}
