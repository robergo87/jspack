package main

import (
    "fmt"
    "os"
    "io"
    "errors"
    "strings"
    "bytes"
    //"path"
    "path/filepath"
    //"encoding/xml"
    "golang.org/x/net/html"
    "encoding/json"
    "os/exec"
    "crypto/md5"
)

func string_between(haystack, begin, end string) string{
    pos_start := strings.Index(haystack, begin)
    if pos_start < 0 {return ""}
    pos_start += len(begin)
    pos_end := strings.Index(haystack, end)
    if pos_end < 0 {return ""}
    return haystack[pos_start:pos_end]
}

func validate_js(content string) error {
    nodeErrors := new(strings.Builder)
    nodecheck := exec.Command("nodejs", "-c")
    nodecheck.Stdin = strings.NewReader("\n"+content)
    nodecheck.Stderr = nodeErrors
    err := nodecheck.Run()
    if err != nil {return errors.New(nodeErrors.String())}
    return nil
}

func validate_html(content string) error {
    tokenizer := html.NewTokenizer(strings.NewReader(content))
    tags := []string{}
    
    for {
        tokenType := tokenizer.Next()
        if tokenType == html.ErrorToken {
            err := tokenizer.Err()
            if err == io.EOF {break}
            return err
        }
        if tokenType == html.StartTagToken {
            token := string(tokenizer.Token().Data)
            tags = append(tags, token)
        }
        if tokenType == html.EndTagToken {
            token := string(tokenizer.Token().Data)
            if len(tags) == 0 {
                return errors.New(
                    fmt.Sprintf("Unexpected closing tag </%s>", token),
                )
            }
            lastTag := tags[len(tags)-1]
            if lastTag != token {
                return errors.New(
                    fmt.Sprintf("Tag mismatch, expected </%s>, got </%s>", lastTag, token),
                )
            }
            tags = tags[0:len(tags)-1]
        }
    }
    if len(tags) > 0 {
        lastTag := tags[len(tags)-1]
        return errors.New(
            fmt.Sprintf("Unclosed tag </%s>", lastTag),
        )
    }
    return nil
}

func validate_css(content string) error {
    return nil
}

type Parser struct {
    filepath string
    rawcontent []byte
    loaded bool 
}

type ParserInterface interface {
    GetPath() string
    Load() error
    Render() (string, error)
}


func NewParser(fp string) ParserInterface {
    ext := filepath.Ext(fp)
    if ext == ".js" {return NewJSParser(fp)}
    if ext == ".css" {return NewCSSParser(fp)}
    if ext == ".html" {return NewMixedParser(fp)}
    return nil
}

func (self *Parser) LoadFileContent() error {
    filecontent, err := os.ReadFile(self.filepath)
    if err != nil {return err}    
    self.rawcontent = filecontent
    self.loaded = true
    return nil
}


func (self *Parser) GetPath() string {
    return self.filepath
}

func (self *Parser) Load() error {
    return self.LoadFileContent()
}

func (self Parser) Render() (string, error) {
    if !self.loaded {
        return "", errors.New("File not loaded")
    }
    hash := md5.Sum([]byte(self.filepath))
    json_content, err := json.Marshal(string(self.rawcontent))
    if err != nil {return "", err}
    return fmt.Sprintf("FILE_CONTENT[\"%x\"] = %s\n", hash, json_content), nil
}


type CSSParser struct {
    Parser    
}
func NewCSSParser(filepath string) *CSSParser {
    obj := CSSParser{}
    obj.filepath = filepath
    return &obj
}

type HTMLParser struct {
    Parser    
}
func NewHTMLParser(filepath string) *HTMLParser {
    obj := HTMLParser{}
    obj.filepath = filepath
    return &obj
}

type TextParser struct {
    Parser    
}
func NewTextParser(filepath string) *TextParser {
    obj := TextParser{}
    obj.filepath = filepath
    return &obj
}

type JSParser struct {
    Parser    
}
func NewJSParser(filepath string) *JSParser {
    obj := JSParser{}
    obj.filepath = filepath
    return &obj
}
func (self *JSParser) Load() error {
    err := self.LoadFileContent()
    if err != nil {return err}
    err = validate_js(string(self.rawcontent))
    return err
}

func (self *JSParser) Render() (string, error) {
    if !self.loaded {
        return "", errors.New("File not loaded")
    }
    return string(self.rawcontent), nil
}



type MixedParser struct {
    Parser
    script string
    template string
    style string
}
func NewMixedParser(filepath string) *MixedParser {
    obj := MixedParser{}
    obj.filepath = filepath
    return &obj
}

func (self *MixedParser) Load() error {
    err := self.LoadFileContent()
    if err != nil {return err}
    
    tokenizer := html.NewTokenizer(bytes.NewReader(self.rawcontent))
    currentTag := ""
    currentTagCount := 0
    currentTagLine := 0
    currentContent := ""
    lineno := 1
    for {
        tokenType := tokenizer.Next()
        if tokenType == html.ErrorToken {
            err := tokenizer.Err()
            if err == io.EOF {break}
            return err
        }
        token := string(tokenizer.Raw())
        lineno += strings.Count(token, "\n")

        if tokenType == html.StartTagToken {
            if currentTagCount == 0 {
                if token == "<script>" || token == "<style>" || token == "<template>" {
                    currentTag = token[1:len(token)-1]
                    currentContent = ""
                    currentTagCount += 1
                    currentTagLine = lineno
                    continue
                } else {
                    return errors.New(fmt.Sprintf("Unsupported opening tag %s at line %i", token, lineno))
                }
            }
            if currentTagCount > 0 && token == "<"+currentTag+">" {
                currentTagCount += 1
            }
        }
        if tokenType == html.EndTagToken {
            if currentTagCount == 0 {
                if token != "</script>" && token != "</style>" && token != "</template>" {
                    return errors.New(fmt.Sprintf("Unsupported closing tag %s at line %x", token, lineno))
                } else {
                    return errors.New(fmt.Sprintf("Unmatched closing tag %s at line %d", token, lineno))
                }
            }
            if token == "</"+currentTag+">" {
                currentTagCount -= 1
                if currentTagCount == 0 {
                    currentTag = ""
                    if token == "</script>" {self.script = currentContent}
                    if token == "</template>" {self.template = currentContent}
                    if token == "</style>" {self.style = currentContent}
                    continue
                }
            }
        }
        if currentTagCount > 0 {
            currentContent += token
        }
    }
    if currentTagCount > 0 {
        return errors.New(fmt.Sprintf("Unmatched opening tag %s from line %d", currentTag, currentTagLine))
    }
    err = validate_js(self.script)
    if err != nil {return err}
    err = validate_html(self.template)
    if err != nil {return err}
    return nil
}

func (self *MixedParser) Render() (string, error) {
    if !self.loaded {
        return "", errors.New("File not loaded")
    }
    retval := ""
    if len(self.style) > 0 {
        hash_css := md5.Sum([]byte(self.filepath+".css"))
        css_content, err := json.Marshal(self.style)
        if err != nil {return "", err}
        retval += fmt.Sprintf("FILE_CONTENT[\"%x\"] = %s\n", hash_css, css_content)
    }
    
    if len(self.template) > 0 {
        hash_html := md5.Sum([]byte(self.filepath+".html"))
        xml_content, err := json.Marshal(self.template)
        if err != nil {return "", err}
        retval += fmt.Sprintf("FILE_CONTENT[\"%x\"] = %s\n", hash_html, xml_content)
    }

    if len(self.script) > 0 {
        retval += fmt.Sprintf("%s", self.script)
    }
    return retval, nil
}

