module Data.Note exposing (..)

import Url.Parser as P exposing (Parser)


type Slug
    = Slug String


urlSlugParser : Parser (Slug -> a) a
urlSlugParser =
    P.custom "SLUG" (\s -> Just (Slug s))
