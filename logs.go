package main

import (
    "log"
)


const TERM_RESET  = "\033[0m"
const TERM_RED    = "\033[31m"
const TERM_GREEN  = "\033[32m"
const TERM_YELLOW = "\033[33m"
const TERM_BLUE   = "\033[34m"
const TERM_PURPLE = "\033[35m"
const TERM_CYAN   = "\033[36m"
const TERM_GRAY   = "\033[37m"
const TERM_WHITE  = "\033[97m"

func log_color(color string, v ...any ) {
    log_msg := []any{color}
    log_msg = append(log_msg, v...)
    log_msg = append(log_msg, TERM_RESET)
    log.Println( log_msg... )
}

func log_notice(v ...any ) {
    log_color(TERM_BLUE, v...)
}

func log_warning(v ...any ) {
    log_color(TERM_YELLOW, v...)
}

func log_error(v ...any ) {
    log_color(TERM_RED, v...)
}

func log_success(v ...any ) {
    log_color(TERM_GREEN, v...)
}

