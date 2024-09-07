module [
    Request,
    Response,
    Method,
    parseFormUrlEncoded,
    methodToStr,
]

import Webserver

Request : Webserver.Request
Response : Webserver.Response
Method : Webserver.RequestMethod

# The below is coppied from basic-webserver

## Parse URL-encoded form values (`todo=foo&status=bar`) into a Dict (`("todo", "foo"), ("status", "bar")`).
##
## ```
## expect
##     bytes = Str.toUtf8 "todo=foo&status=bar"
##     parsed = parseFormUrlEncoded bytes |> Result.withDefault (Dict.empty {})
##
##     Dict.toList parsed == [("todo", "foo"), ("status", "bar")]
## ```
parseFormUrlEncoded : List U8 -> Result (Dict Str Str) [BadUtf8]
parseFormUrlEncoded = \bytes ->

    chainUtf8 = \bytesList, tryFun -> Str.fromUtf8 bytesList |> mapUtf8Err |> Result.try tryFun

    # simplify `BadUtf8 Utf8ByteProblem ...` error
    mapUtf8Err = \err -> err |> Result.mapErr \_ -> BadUtf8

    parse = \bytesRemaining, state, key, chomped, dict ->
        tail = List.dropFirst bytesRemaining 1

        when bytesRemaining is
            [] if List.isEmpty chomped -> dict |> Ok
            [] ->
                # chomped last value
                key
                |> chainUtf8 \keyStr ->
                    chomped
                    |> chainUtf8 \valueStr ->
                        Dict.insert dict keyStr valueStr |> Ok

            ['=', ..] -> parse tail ParsingValue chomped [] dict # put chomped into key
            ['&', ..] ->
                key
                |> chainUtf8 \keyStr ->
                    chomped
                    |> chainUtf8 \valueStr ->
                        parse tail ParsingKey [] [] (Dict.insert dict keyStr valueStr)

            ['%', secondByte, thirdByte, ..] ->
                hex = Num.toU8 (hexBytesToU32 [secondByte, thirdByte])

                parse (List.dropFirst tail 2) state key (List.append chomped hex) dict

            [firstByte, ..] -> parse tail state key (List.append chomped firstByte) dict

    parse bytes ParsingKey [] [] (Dict.empty {})

hexBytesToU32 : List U8 -> U32
hexBytesToU32 = \bytes ->
    bytes
    |> List.reverse
    |> List.walkWithIndex 0 \accum, byte, i -> accum + (Num.powInt 16 (Num.toU32 i)) * (hexToDec byte)
    |> Num.toU32

hexToDec : U8 -> U32
hexToDec = \byte ->
    when byte is
        '0' -> 0
        '1' -> 1
        '2' -> 2
        '3' -> 3
        '4' -> 4
        '5' -> 5
        '6' -> 6
        '7' -> 7
        '8' -> 8
        '9' -> 9
        'A' -> 10
        'B' -> 11
        'C' -> 12
        'D' -> 13
        'E' -> 14
        'F' -> 15
        _ -> crash "Impossible error: the `when` block I'm in should have matched before reaching the catch-all `_`."

methodToStr : Method -> Str
methodToStr = \method ->
    when method is
        Options -> "Options"
        Get -> "Get"
        Post _ -> "Post"
        Put _ -> "Put"
        Delete _ -> "Delete"
        Head -> "Head"
        Trace -> "Trace"
        Connect -> "Connect"
        Patch _ -> "Patch"
