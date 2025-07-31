module Constants exposing (expirationOptions)


expirationOptions : List { text : String, value : Int }
expirationOptions =
    [ { text = "Never expires (default)", value = 0 }
    , { text = "1 hour", value = 60 * 60 * 1000 }
    , { text = "12 hours", value = 12 * 60 * 60 * 1000 }
    , { text = "1 day", value = 24 * 60 * 60 * 1000 }
    , { text = "3 days", value = 3 * 24 * 60 * 60 * 1000 }
    , { text = "7 days", value = 7 * 24 * 60 * 60 * 1000 }
    ]
