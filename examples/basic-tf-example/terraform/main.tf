resource "local_file" "myvar" {
    content  = var.myvar
    filename = "${path.module}/myvar"
}