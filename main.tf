resource "tinyfaas_function" "blubb-func" {
    name = "blubb"
    num_threads = 3
    tarball_path = "./blubb.tar"
    environment = {
        KEY = "value"
    }
}
