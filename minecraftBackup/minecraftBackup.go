package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

/*
# プログラム仕様書
1. 外部コマンドtarが使えるか確認。
2.バックアップ対象一覧の確認
	backup.txtに記載された複数行のディレクトリ名を読み取る。
3.マインクラフトのセーブデータ場所に移動。
	外出ししたテキストファイルからの読み取り？
4.上記1で読み込んだセーブディレクトリを順にtarで固める。
5.固めたtarファイルを順にデスクトップに移動する。
6.作業ディレクトリに戻る。

## 方針
第1弾：外部コマンドによるアーカイブでのバックアップ	←今回。
第2弾：自前(Golang)によるアーカイブでのバックアップ
*/

// readBackupFile バックアップ対象ディレクトリ一覧ファイル
const readBackupFile = "backup.txt"

// saveDataDirFile セーブデータディレクトリ
const saveDataDirFile = "saveDir.txt"

func main() {
	out, err := exec.Command("tar", "--version").Output()
	if err != nil {
		log.Fatal("tarコマンド失敗", err)
	} else if out == nil {
		log.Fatal("tarコマンドない？", err)
	}
	fmt.Print("tarのバージョン：", string(out))

	readFile, errBackFile := os.Open(readBackupFile)
	if errBackFile != nil {
		//log.Fatal("バックアップファイルが存在しない", err)
		// セーブデータ用ディレクトリをすべてバックアップ実施。
	}
	defer readFile.Close()

	pathFile, errPathFile := os.OpenFile(saveDataDirFile, os.O_RDWR|os.O_CREATE, 0666)
	var dirPath string = ""        // セーブデータ用のPath
	var dirDesktopPath string = "" // セーブデータのアーカイブ先のPath(要はデスクトップ)
	if errPathFile != nil {
		log.Fatal("セーブデータ用のPathが存在しない。", err)
	}
	linner := bufio.NewScanner(pathFile)
	for linner.Scan() {
		dirPath = linner.Text()
		if linner.Text() == "" {
			// 空行はスキップ
			continue
		} else if dirTrueFalse, err := os.Stat(dirPath); os.IsNotExist(err) || !dirTrueFalse.IsDir() {
			// ディレクトリではないためスキップ
			dirPath = ""
			continue
		} else {
			// 存在するセーブ先ディレクトリのため、これを使う。
			//fmt.Println(dirTrueFalse.Name())
			break
		}
	}
	home, err := os.UserHomeDir()
	dirDesktopPath = home + "/Desktop/" // セーブデータアーカイブの保存先。
	if err != nil {
		log.Fatal("ホームディレクトリ取得失敗", err)
		return
	}
	if dirPath == "" {
		// デフォルトPathをセーブデータとする。
		if runtime.GOOS == "windows" {
			// Java版のみ？
			dirPath = home + "/AppData/Roaming/.minecraft/saves/"
			// 以下が統合版？
			//dirPath = home + "/AppData/Local/Packages/Microsoft.MinecraftUWP_8wekyb3d8bbwe/LocalState/games/com.mojang"
		} else {
			// MacOS(Java版)限定？
			dirPath = home + "/Library/Application Support/minecraft/saves/"
		}
		fmt.Println(dirPath)
		if dirTrueFalse, err := os.Stat(dirPath); os.IsNotExist(err) || !dirTrueFalse.IsDir() {
			// 参考URL：https://qiita.com/hnakamur/items/848097aad846d40ae84b
		} else {
			// 存在するディレクトリのため、これを使う。
			//fmt.Println("デフォルトセーブディレクトリ", dirPath)
		}
	}

	// 現在地の保存
	prev, err := filepath.Abs(".")
	if err != nil {
		return // ERROR
	}
	defer os.Chdir(prev)

	// ディレクトリ移動
	os.Chdir(dirPath)

	// セーブデータ用のディレクトリ名取得
	if readFile == nil {
		// セーブデータ用ディレクトリをすべてバックアップ実施。
		saveFiles, _ := ioutil.ReadDir("./")
		for _, saveDir := range saveFiles {
			if saveDir.IsDir() {
				// 存在するディレクトリのため、これを使う。
				//fmt.Println("セーブディレクトリ", saveDir.Name())
				tarName := saveDir.Name() + ".tar.gz"
				_, err := exec.Command("tar", "zcvf", tarName, saveDir.Name()).Output()
				if err == nil {
					// デスクトップに移動。
					err = os.Rename(tarName, dirDesktopPath+tarName)
					if err != nil {
						fmt.Println("セーブデータをデスクトップに移動失敗", err)
					}
				} else {
					fmt.Println("セーブ対象用ディレクトリが存在しない", err)
				}
			}
			// fmt.Println("セーブ対象外ファイル", saveDir)
		}
	} else {
		// ユーザが絞り込んだバックアップ対象のセーブデータのみバックアップ実施。
		scanner := bufio.NewScanner(readFile)
		for scanner.Scan() {
			saveDir := scanner.Text()
			if saveDir == "" {
				//				fmt.Println("セーブデータ用のディレクトリ名取得だが、空行なのでスキップ")
				continue
			} else if dirTrueFalse, err := os.Stat(saveDir); os.IsNotExist(err) || !dirTrueFalse.IsDir() {
				// ディレクトリではない。
				fmt.Printf("指定のディレクトリ(\"%s\")が存在しない(もしくは、セーブデータPathの間違い)。\n", saveDir)
				continue
			} else {
				// 存在するディレクトリのため、これを使う。
				//fmt.Println("セーブディレクトリ配下：", saveDir)
				tarName := saveDir + ".tar.gz"
				_, err := exec.Command("tar", "zcvf", tarName, saveDir).Output()
				if err == nil {
					// デスクトップに移動。
					err = os.Rename(tarName, dirDesktopPath+tarName)
					if err != nil {
						fmt.Println("セーブデータをデスクトップに移動失敗", err)
					}
				} else {
					log.Fatal("存在しないセーブディレクトリ", saveDir, err)
				}
				break
			}
		}
	}
	fmt.Println("Mincecraftセーブデータプログラム実行終了")
}
