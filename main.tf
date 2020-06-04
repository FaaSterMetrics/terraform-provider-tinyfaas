resource "tinyfaas_function" "blubb-func" {
    address = "127.0.0.1"
    name = "blubb"
    num_threads = 3
    tarball_path = "./blubb.tar"
}
