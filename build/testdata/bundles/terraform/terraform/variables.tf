variable "file_contents" {
  description = "Contents of the file 'foo'"
  default     = "bar"
}
variable "map_var" {
  description = "Map variable"
  type        = map(string)
  default     = { foo = "bar" }
}

variable "array_var" {
  description = "Array Variable"
  type        = list(any)
  default     = ["mylist"]
}

variable "boolean_var" {
  description = "Boolean Variable"
  type        = bool
  default     = false
}

variable "number_var" {
  description = "Number Variable"
  type        = number
  default     = 0
}
variable "complex_object_var" {
  description = "Object variable"
  type        = object({
    top_value = string
    nested_object = object({
      internal_value = string
    })
  })
}
