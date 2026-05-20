import 'dart:io';
import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:nexusacg/core/constants/app_constants.dart';
import 'package:nexusacg/core/network/upload_result.dart';

class ApiClient {
  static final ApiClient _instance = ApiClient._internal();
  factory ApiClient() => _instance;
  ApiClient._internal();

  late Dio _dio;
  String? _accessToken;

  Future<void> init() async {
    _dio = Dio(BaseOptions(
      baseUrl: AppConstants.apiBaseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 30),
      headers: {'Content-Type': 'application/json'},
      validateStatus: (status) => true, // Never throw on HTTP errors
    ));

    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) {
        if (_accessToken != null) {
          options.headers['Authorization'] = 'Bearer $_accessToken';
        }
        handler.next(options);
      },
      onError: (error, handler) {
        if (error.response?.statusCode == 401) {
          // Token expired, could trigger refresh flow
        }
        handler.next(error);
      },
    ));

    final prefs = await SharedPreferences.getInstance();
    _accessToken = prefs.getString('access_token');
  }

  set accessToken(String? token) {
    _accessToken = token;
    SharedPreferences.getInstance().then((prefs) {
      if (token != null) {
        prefs.setString('access_token', token);
      } else {
        prefs.remove('access_token');
      }
    });
  }

  Dio get dio => _dio;

  Future<Response<T>> get<T>(String path, {Map<String, dynamic>? queryParameters}) async {
    return _dio.get<T>(path, queryParameters: queryParameters);
  }

  Future<Response<T>> post<T>(String path, {dynamic data}) async {
    return _dio.post<T>(path, data: data);
  }

  Future<Response<T>> delete<T>(String path) async {
    return _dio.delete<T>(path);
  }

  Future<Response<T>> put<T>(String path, {dynamic data}) async {
    return _dio.put<T>(path, data: data);
  }

  Future<UploadResult> uploadVideo(File file, {void Function(int, int)? onProgress}) async {
    final formData = FormData.fromMap({
      'file': await MultipartFile.fromFile(file.path, contentType: DioMediaType.parse('video/mp4')),
    });
    final response = await _dio.post(
      '/upload/video',
      data: formData,
      options: Options(
        headers: {'Content-Type': 'multipart/form-data'},
        sendTimeout: const Duration(minutes: 2),
      ),
      onSendProgress: onProgress,
    );
    final data = response.data;
    if (data is Map && data['code'] == 0 && data['data'] != null) {
      final inner = data['data'] as Map;
      final url = inner['url'] as String?;
      if (url != null) return UploadResult(url: url);
    }
    if (data is Map && data['message'] != null) {
      return UploadResult(error: data['message'].toString());
    }
    return UploadResult(error: '服务器未返回URL (HTTP ${response.statusCode})');
  }

  Future<String?> uploadImage(File file) async {
    final formData = FormData.fromMap({
      'file': await MultipartFile.fromFile(file.path),
    });
    final response = await _dio.post(
      '/upload',
      data: formData,
      options: Options(headers: {'Content-Type': 'multipart/form-data'}),
    );
    final data = response.data;
    if (data is Map && data['code'] == 0 && data['data'] != null) {
      final inner = data['data'] as Map;
      return inner['url'] as String?;
    }
    return null;
  }
}
