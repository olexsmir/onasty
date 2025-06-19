module Auth exposing (User, onPageLoad, viewCustomPage)

import Auth.Action
import Auth.User
import Data.Credentials exposing (Credentials)
import Dict
import Route exposing (Route)
import Route.Path
import Shared
import View exposing (View)


type alias User =
    Auth.User.User


{-| Called before an auth-only page is loaded.
-}
onPageLoad : Shared.Model -> Route () -> Auth.Action.Action User
onPageLoad shared _ =
    case shared.user of
        Auth.User.NotSignedIn ->
            Auth.Action.pushRoute
                { path = Route.Path.Auth
                , query = Dict.empty
                , hash = Nothing
                }

        Auth.User.SignedIn credentials ->
            Auth.Action.loadPageWithUser credentials


{-| Renders whenever `Auth.Action.loadCustomPage` is returned from `onPageLoad`.
-}
viewCustomPage : Shared.Model -> Route () -> View Never
viewCustomPage _ _ =
    View.fromString "Loading..."
