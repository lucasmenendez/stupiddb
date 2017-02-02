PACKAGE DOCUMENTATION

package encoder
    import "./encoder/"


FUNCTIONS

func Bool(data bool) (interface{}, error)
    Serialize Boolean. If something fails returns a 'DBError' with info
    message.

func Float(data float64) (interface{}, error)
    Build Float with format defined by type check correct data provided. If
    something fails returns a 'DBError' with info message.

func Int(data int64) (interface{}, error)
    Build Integer with format defined by type check correct data provided.
    If something fails returns a 'DBError' with info message.

func String(data string, size int) (interface{}, error)
    Build String with format defined by size and according offset and check
    data provided length. If something fails returns a 'DBError' with info
    message.

