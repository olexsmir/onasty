module Auth exposing (User, onPageLoad, viewCustomPage)

import Auth.Action
import Dict
import Route exposing (Route)
import Route.Path
import Shared
import View exposing (View)


type alias User =
    { accessToken : String
    , refreshToken : String
    }


{-| Called before an auth-only page is loaded.
-}
onPageLoad : Shared.Model -> Route () -> Auth.Action.Action User
onPageLoad shared _ =
    case shared.credentials of
        Just credentials ->
            Auth.Action.loadPageWithUser
                { accessToken = credentials.accessToken
                , refreshToken = credentials.refreshToken
                }

        _ ->
            Auth.Action.pushRoute
                { path = Route.Path.Auth
                , query = Dict.empty
                , hash = Nothing
                }


{-| Renders whenever `Auth.Action.loadCustomPage` is returned from `onPageLoad`.
-}
viewCustomPage : Shared.Model -> Route () -> View Never
viewCustomPage _ _ =
    View.fromString "Loading..."
