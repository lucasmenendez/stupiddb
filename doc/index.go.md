PACKAGE DOCUMENTATION

package index
    import "./index/"


FUNCTIONS

func Get(index_path string) ([]*Index, error)
    Returns a Index's pointer from table location provided. First check if
    table has any index and then returns all index found with its
    attributues and content. If something fails returns a 'DBError' with
    info message.

func New(location, table, column string) error
    Create new index for table column provided on its location. If this
    index its the first, create folder tree to store its table index. If
    something fails returns a 'DBError' with info message.

TYPES

type Index struct {
    Column   string
    Location string
    Content  []string
    Mutex    *sync.Mutex
}
    Define index struct atributtes that contains column name, location path,
    content and its mutex.

func (index *Index) Append(content string) error
    Insert new record to selected index on temp memory and commit the change
    on associated file. Lock file while commiting the change. If something
    fails returns a 'DBError' with info message.

func (index *Index) Delete(content string) error
    Delete record from temp memory and form index file. Lock file while
    committing the change. If something fails returns a 'DBError' with info
    message.

func (index *Index) Exist(needle string) bool
    Check if record exists on selected index by content provided encoded
    previously. Return true in record exists.

func (index *Index) Find(needle string) (int, error)
    Returns record line number location on selected index by encoded
    content. If something fails, returns a 'DBError' with info message.

func (index *Index) Remove() error
    Remove column index file if exists. If something fails returns a
    'DBError' with info message.
