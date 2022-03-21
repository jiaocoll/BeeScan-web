package runner

import (
	"Beescan/core/poc/nuclei"
	"Beescan/core/poc/xray"
	"Beescan/core/util"
	"errors"
	"fmt"
	"github.com/karrick/godirwalk"
	"github.com/projectdiscovery/nuclei/v2/pkg/templates"
	"github.com/remeh/sizedwaitgroup"
	"log"
	"path/filepath"
	"strings"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/2/23
程序功能：poc模块
*/

// LoadTemplates 加载poc
func (r *Runner) LoadTemplates(f string) []string {
	var allTemplates []string
	// resolve and convert relative to absolute path
	var absPath string
	var err error

	if strings.Contains(f, "*") {
		dirs := strings.Split(f, "/")
		priorDir := strings.Join(dirs[:len(dirs)-1], "/")
		absPath, err = util.ResolvePathIfRelative(priorDir)
		absPath += "/" + dirs[len(dirs)-1]
	} else {
		// resolve and convert relative to absolute path
		absPath, err = util.ResolvePathIfRelative(f)
	}

	if err != nil {
		log.Printf("Could not find template file '%s': %s\n", f, err)
		return allTemplates
	}
	// Template input includes a wildcard
	if strings.Contains(absPath, "*") {
		var matches []string
		matches, err = filepath.Glob(absPath)

		if err != nil {
			log.Printf("Wildcard found, but unable to glob '%s': %s\n", absPath, err)
			return allTemplates
		}

		// couldn't find templates in directory
		if len(matches) == 0 {
			log.Printf("Error, no templates were found with '%s'.\n", absPath)
			return allTemplates
		} else {
			log.Printf("Identified %d templates\n", len(matches))
		}

		for _, match := range matches {
			allTemplates = append(allTemplates, match)
		}
	} else {
		isFile, err := util.IsFilePath(absPath)
		if err != nil {
			log.Printf("Could not stat '%s': %s\n", absPath, err)
			return allTemplates
		}
		if isFile {
			allTemplates = append(allTemplates, absPath)
		} else {
			// 是一个目录
			// Recursively walk down the Templates directory and run all the template file checks
			err := util.DirectoryWalker(
				absPath,
				func(path string, d *godirwalk.Dirent) error {
					if !d.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
						allTemplates = append(allTemplates, path)
					}
					return nil
				},
			)
			// directory couldn't be walked
			if err != nil {
				log.Printf("Could not find templates in directory '%s': %s\n", absPath, err)
				return allTemplates
			}

		}
	}

	return allTemplates
}

// ParseTemplates 解析文件并返回指定格式
func (r *Runner) ParseTemplate(f string) (interface{}, error) {
	// check if it's a workflow
	templateXray, errXray := xray.ParsePocFile(f)
	if errXray == nil {
		return templateXray, nil
	}
	templatesNuclei, errNuclei := r.nuclei.ParsePocFile(f)
	if errNuclei == nil {
		return templatesNuclei, nil
	}
	return nil, errors.New(fmt.Sprintf("unknown error occurred:%s", f))
}

// ParsePocs 解析poc
func (r *Runner) ParsePocs() int {
	var pocTemplates []string

	pocTemplates = append(pocTemplates, r.LoadTemplates("data/")...)

	send := 0
	var loadFaild []string
	for _, match := range pocTemplates {
		t, err := r.ParseTemplate(match)
		switch t.(type) {
		case *xray.Poc:
			poc := t.(*xray.Poc)
			err2 := poc.Check()
			if err2 != nil {
				loadFaild = append(loadFaild, fmt.Sprintf("Xray PoC加载失败 %s 作者:%s 失败原因:%s", match, poc.Detail.Author, err2.Error()))
				break
			}

			r.xrayPocs = append(r.xrayPocs, poc)
			send += 1
		case *templates.Template:
			poc := t.(*templates.Template)
			r.nucleiPoCs = append(r.nucleiPoCs, poc)
			send += poc.TotalRequests
		default:
			log.Printf("Could not parse file '%s': %v", match, err)
		}
	}
	return send
}

// RunPocs 运行poc
func (r *Runner) RunPocs(target string, output chan PocResult) {
	var pocs []interface{}
	for _, poc := range r.xrayPocs {
		pocs = append(pocs, poc)
	}
	for _, poc := range r.nucleiPoCs {
		pocs = append(pocs, poc)
	}
	if len(pocs) == 0 {
		return
	}

	wg := sizedwaitgroup.New(5)
	for _, poc := range pocs {
		wg.Add()
		go func(pocInterface interface{}) {
			defer wg.Done()
			var vbool bool = false
			var err error

			switch pocInterface.(type) {
			case *xray.Poc:
				poc := pocInterface.(*xray.Poc)
				vbool, err = poc.Execute(target, r.hp)
				if err != nil {
					log.Println(err)
					return
				}

				pocOutput := PocResult{
					ID:             target + "-" + poc.Name,
					TaskName:       r.taskname,
					Target:         target,
					PocName:        poc.Name,
					PocLink:        poc.Detail.Links,
					PocAuthor:      poc.Detail.Author,
					PocDescription: poc.Detail.Description,
					LastTime:       time.Now().Format("2006-01-02 15:04:05"),
				}

				if vbool {
					output <- pocOutput
				}
			case *templates.Template:
				poc := pocInterface.(*templates.Template)
				results, err := nuclei.ExecuteNucleiPoc(target, poc)
				if err != nil || results == nil {
					return
				}
				for _, ret := range results {
					ret2 := ret
					nameInterFace := poc.Info.Name
					authorInterFace := poc.Info.Authors.String()
					name := fmt.Sprintf("%s", nameInterFace)
					author := fmt.Sprintf("%s", authorInterFace)

					pocOutput := PocResult{
						ID:             target + "-" + name,
						TaskName:       r.taskname,
						Target:         target,
						PocName:        name,
						PocLink:        []string{},
						PocAuthor:      author,
						PocDescription: ret2,
						LastTime:       time.Now().Format("2006-01-02 15:04:05"),
					}
					output <- pocOutput
				}
			}
		}(poc)
	}
	wg.Wait()
}
