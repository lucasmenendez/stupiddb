PACKAGE DOCUMENTATION

package types
    import "./types/"


TYPES

type Type struct {
    Alias     string
    Indexable bool
    Size      int
    Content   interface{}
}
    Type contains column attributes that defines data type, length and
    provided content access.

func Bool() Type
    Returns empty Bool type struct.

func Float() Type
    Returns empty Float type struct.

func Int(indexable bool) Type
    Returns empty Int type struct.

func String(size int, indexable bool) Type
    Returns empty String type struct with length provided.

func (data *Type) Decoder() error
    Fill Type content typed according with associated data structure. If
    something was wrong returns a 'DBError', else nil.

func (data *Type) Empty() bool
    Returns if current type is empty

func (data *Type) Encoder() error
    Fill Type content with encoded typed representation data content
    according to associate data structure. If something was wrong returns a
    'DBError', else nil.

