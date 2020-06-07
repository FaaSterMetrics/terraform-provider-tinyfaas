resource "tinyfaas_function" "blubb-func" {
    name = "blubb"
    num_threads = 3
    zip_path = "./blubb.zip"
    environment = {
        KEY = "value"
    }
}
