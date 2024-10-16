module Viewer exposing (..)

import Api exposing (Cred)


{-| The logged-in user currently viewing this page. It stores enough data to
be able to render the menu bar, along with Cred so it's impossible to have a
Viewer if you aren't logged in.
-}
type Viewer
    = Viewer Cred
