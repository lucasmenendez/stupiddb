PACKAGE DOCUMENTATION

package decoder
    import "./decoder/"


FUNCTIONS

func Bool(data []byte) (interface{}, error)
    After check input data format return casted bool from data provided. If
    something fails returns a 'DBError' with info message.

func Float(data []byte) (interface{}, error)
    After check input data format return casted float from data provided. If
    something fails returns a 'DBError' with info message.

func Int(data []byte) (interface{}, error)
    After check input data format return casted int from data provided. If
    something fails returns a 'DBError' with info message.

func String(data []byte) (interface{}, error)
    Return casted string with 'trim()' from data provided. If something
    fails returns a 'DBError' with info message.
