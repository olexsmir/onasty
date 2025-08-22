module Validators exposing (email, password, passwords)


email : String -> Maybe String
email inp =
    if
        not (String.isEmpty inp)
            && (not (String.contains "@" inp) && not (String.contains "." inp))
    then
        Just "Please enter a valid email address."

    else
        Nothing


password : String -> Maybe String
password passwd =
    if not (String.isEmpty passwd) && String.length passwd < 8 then
        Just "Password must be at least 8 characters long."

    else
        Nothing


passwords : String -> String -> Maybe String
passwords passowrd1 password2 =
    if not (String.isEmpty passowrd1) && passowrd1 /= password2 then
        Just "Passwords do not match."

    else
        Nothing
