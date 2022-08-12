variable "file_contents" {
  description = "Contents of the file 'foo'"
  default     = "bar"
}
variable "map_var" {
  description = "Object variable"
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
