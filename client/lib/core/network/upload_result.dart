class UploadResult {
  final String? url;
  final String? error;
  bool get isSuccess => url != null;

  const UploadResult({this.url, this.error});
}
