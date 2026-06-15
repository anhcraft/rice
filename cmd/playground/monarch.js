function _() {
    return {
        defaultToken: 'invalid',
        tokenPostfix: '.rice',

        keywords: [
            "var", "const", "if", "else", "func", "for", "in", "continue", "break", "return"
        ],

        literals: [
            "true", "false", "null"
        ],

        operators: [
            '=', '>', '<', '!', '==', '<=', '>=', '!=', '&&', '||', '++', '--',
            '+', '-', '*', '/', '%', '...'
        ],

        symbols: /\.\.\.|--|\+\+|&&|\|\||==|!=|>=|<=|[=><!+\-*/(){}\[\].,;]/,

        escapes: /\\(?:[bfnrt\\"']|u[0-9A-Fa-f]{4}|U[0-9A-Fa-f]{8})/,

        tokenizer: {
            root: [
                // Identifiers and keywords
                [/[a-zA-Z_][\w_]*/, {
                    cases: {
                        '@keywords': 'keyword',
                        '@literals': 'keyword.literal',
                        '@default': 'identifier'
                    }
                }],

                // Whitespace and comments
                {include: '@whitespace'},

                // Delimiters and operators
                [/[{}()\[\]]/, '@brackets'],
                [/@symbols/, {
                    cases: {
                        '@operators': 'operator',
                        '@default': ''
                    }
                }],

                // Numbers
                // 1. Floats (must contain a dot or an exponent)
                [/\d*\.\d+([eE][\-+]?\d+)?/, 'number.float'],
                [/\d+[eE][\-+]?\d+/, 'number.float'],
                // 2. Integers
                [/\d+/, 'number'],

                // Delimiters (comma, semicolon)
                [/[;,]/, 'delimiter'],

                // Strings
                [/"/, 'string', '@string'],
            ],

            whitespace: [
                [/[ \t\r\n]+/, ''], // White space
                [/#.*$/, 'comment'], // Comments start with #
            ],

            string: [
                [/[^\\"]+/, 'string'],
                [/@escapes/, 'string.escape'],
                [/\\./, 'string.escape.invalid'],
                [/"/, 'string', '@pop']
            ],
        },
    };
}
