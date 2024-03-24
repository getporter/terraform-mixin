resource "local_file" "foo" {
  content  = var.infra1_var
  filename = "${path.module}/infra1"
}
