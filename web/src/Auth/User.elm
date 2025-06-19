module Auth.User exposing (SignInStatus(..), User)


type alias User =
    { accessToken : String
    , refreshToken : String
    }


type SignInStatus
    = SignedIn User
    | NotSignedIn
