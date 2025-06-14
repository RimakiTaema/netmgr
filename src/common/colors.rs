pub struct Colors {
    pub red: &'static str,
    pub green: &'static str,
    pub yellow: &'static str,
    pub blue: &'static str,
    pub purple: &'static str,
    pub cyan: &'static str,
    pub white: &'static str,
    pub gray: &'static str,
    pub reset: &'static str,
}

pub const COLORS: Colors = Colors {
    red: "\x1b[0;31m",
    green: "\x1b[0;32m",
    yellow: "\x1b[1;33m",
    blue: "\x1b[0;34m",
    purple: "\x1b[0;35m",
    cyan: "\x1b[0;36m",
    white: "\x1b[1;37m",
    gray: "\x1b[0;37m",
    reset: "\x1b[0m",
};
