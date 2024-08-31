### Notes
- uses `gin-gonic`, to address state value validation, JSON marshal/unmarshal, and path parameter parsing, out-of-the-box.
- no caps on how many risk items can be created in total, nor on the number of items that can be retrieved on a single GET.  Pagination for GET /risks is considered out-of-scope.
