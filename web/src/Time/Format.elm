module Time.Format exposing (toString)

import DateFormat
import Time exposing (Posix, Zone)


{-| Formats a given `Posix` time and `Zone` into a human-readable string.

    toString zone posix
        > "July 2nd, 2025 21:05"

-}
toString : Zone -> Posix -> String
toString =
    DateFormat.format
        [ DateFormat.monthNameFull
        , DateFormat.text " "
        , DateFormat.dayOfMonthSuffix
        , DateFormat.text ", "
        , DateFormat.yearNumber
        , DateFormat.text " "
        , DateFormat.hourMilitaryFromOneNumber
        , DateFormat.text ":"
        , DateFormat.minuteFixed
        ]
