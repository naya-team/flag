package flag

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Eraser is function to remove flagger on code
// this code required go import
func Eraser(_flags []string, rootPath string, conn *sql.DB) {

	// this regex to get flag name on code
	// example code:
	// flagger.IsEnable("promo_segmentation_buyer")
	// then this regex will get promo_segmentation_buyer
	re := regexp.MustCompile(`"([^"]+)"`)

	// this regex for delete flagger on middleware
	reMiddleware := regexp.MustCompile(`flag.ReleaseFlag\("[^"]+",\s*([^)]+)\)`)

	// mapping releaseFlags to map
	var (
		flags = make(map[string]bool)
		// dbOrm          *db.GormDB
		isSpecificFlag bool
	)

	// delete spesific flag
	if len(_flags) > 0 {
		for _, flag := range _flags {
			flags[flag] = true
		}
		isSpecificFlag = true
	} else {
		// open connection to db

		// get flag that is enabled for more than 1 month
		// var releaseFlags []ReleaseFlagModel

		oneMonthAgo := time.Now().AddDate(0, -1, 0)

		// select * from release_flags where is_enable = true and enabled_at < oneMonthAgo
		// if error then panic

		rows, err := conn.Query("SELECT flag FROM release_flags WHERE is_enable = ? AND enabled_at < ?", true, oneMonthAgo)
		if err != nil {
			panic(err)
		}

		for rows.Next() {
			var _flag string
			if err := rows.Scan(&_flag); err != nil {
				panic(err)
			}

			flags[_flag] = true
			_flags = append(_flags, _flag)
		}
	}

	// walk on directory and only get file with .go extension
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, _ error) error {
		if filepath.Ext(path) != ".go" {
			return nil
		}

		// open original file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// create temp file
		tempFile, err := os.CreateTemp(".", "tmp-file")
		if err != nil {
			return err
		}
		defer tempFile.Close()

		scanner := bufio.NewScanner(file)

		// remove flag is true when flagger.IsEnable or flagger.ReleaseFlag is found
		// keepCode is false when else is found or code that need to be remove found
		var (
			removeFlag, isNegation bool
			keepCode               = true
		)

		for scanner.Scan() {
			line := scanner.Text()

			if strings.Contains(line, "flag.IsEnable(") || strings.Contains(line, "flag.ReleaseFlag(") {

				if strings.Contains(line, "!flag.IsEnable(") {
					isNegation = true
				}

				// is string contains flag.ReleaseFlag, or found flag on middleware then remove flag
				// example code:
				// router.GET("/", flag.ReleaseFlag("some-Flag"), handler.GetPromo)
				// then should be:
				// router.GET("/", handler.GetPromo)
				if strings.Contains(line, "flag.ReleaseFlag(") {
					line = reMiddleware.ReplaceAllString(line, "$1")
					if _, err := tempFile.WriteString(line + "\n"); err != nil {
						return err
					}
				} else {
					match := re.FindStringSubmatch(line)
					if len(match) >= 2 {
						flag := match[1]
						// check flag enable
						if _, ok := flags[flag]; ok {
							removeFlag = true
							if isNegation {
								keepCode = false
							}
						}
					}
					continue
				}

			} else if removeFlag {
				// remove flagger on code when not negation
				if !isNegation && strings.Contains(line, "else") {
					keepCode = false
					continue
				} else if !isNegation && !keepCode && strings.Contains(line, "}") {
					removeFlag = false
					keepCode = true
					continue
				} else if !isNegation && !keepCode {
					continue
				}

				// remove true condition when negation
				//  keep the else code
				if isNegation && strings.Contains(line, "else") {
					keepCode = true
					continue
				} else if isNegation && keepCode && strings.Contains(line, "}") {
					removeFlag = false
					keepCode = true
					continue
				} else if isNegation && !keepCode {
					continue
				} else if isNegation && keepCode {
					if _, err := tempFile.WriteString(line + "\n"); err != nil {
						return err
					}
				}

			} else {
				if _, err := tempFile.WriteString(line + "\n"); err != nil {
					return err
				}
			}
		}

		if err := file.Close(); err != nil {
			return err
		}
		// tempFile.Seek is to reset cursor to beginning of file
		if _, err := tempFile.Seek(0, 0); err != nil {
			return err
		}

		tempfileContent, err := io.ReadAll(tempFile)
		if err != nil {
			return err
		}

		// write temp file to original file
		err = os.WriteFile(path, tempfileContent, 0644)
		if err != nil {
			return err
		}

		params := []string{"-w", path}
		if err := exec.Command("goimports", params...).Run(); err != nil {
			fmt.Println("cannot run goimports")
		}

		// remove temp file
		if err := os.Remove(tempFile.Name()); err != nil {
			fmt.Println("cannot remove temp file")
			return err
		}

		fmt.Printf("success remove flagger on file %s\n", path)
		return nil
	})

	if err != nil {
		log.Fatalf("error walking the path %q: %v\n", ".", err)
	}

	if !isSpecificFlag {

		// delete flagger on db
		_, err := conn.Exec("DELETE FROM release_flags WHERE flag in (?)", _flags)
		if err != nil {
			log.Fatalf("error delete flagger on db: %v\n", err)
		}
	}

	fmt.Println("done")

}
