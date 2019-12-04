resource "local_file" "a" {
    content  = var.a
    filename = "${path.module}/a"
}

resource "local_file" "b" {
    content  = var.b
    filename = "${path.module}/b"
}

resource "local_file" "c" {
    content  = var.c
    filename = "${path.module}/c"
}