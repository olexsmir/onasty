module Route exposing (..)

import Data.Note
import Url.Parser as Parser exposing ((</>), Parser)


type Route
    = Home
    | SignIn
    | SignUp
    | Logout
    | Note Data.Note.Slug


parser : Parser (Route -> a) a
parser =
    Parser.oneOf
        [ Parser.map Home Parser.top
        , Parser.map SignIn (Parser.s "sign-in")
        , Parser.map SignUp (Parser.s "sign-up")
        , Parser.map Logout (Parser.s "logout")
        , Parser.map Note (Parser.s "n" </> Data.Note.urlSlugParser)
        ]
