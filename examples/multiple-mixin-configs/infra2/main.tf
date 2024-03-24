resource "local_file" "foo" {
  content  = var.infra2_var
  filename = "${path.module}/infra1"
}
