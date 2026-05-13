import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:nexusacg/core/constants/app_constants.dart';

class ApiClient {
  static final ApiClient _instance = ApiClient._internal();
  factory ApiClient() => _instance;
  ApiClient._internal();

  late Dio _dio;
  String? _accessToken;

  Future<void> init() async {
    _dio = Dio(BaseOptions(
      baseUrl: AppConstants.apiBaseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 10),
      headers: {'Content-Type': 'application/json'},
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
}
