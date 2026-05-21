import 'package:shared_preferences/shared_preferences.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_state.dart';

class AuthBloc extends Bloc<AuthEvent, AuthState> {
  final AuthRepository _repo = AuthRepository();

  AuthBloc() : super(AuthInitial()) {
    on<AuthLoginRequested>(_onLogin);
    on<AuthRegisterRequested>(_onRegister);
    on<AuthSmsLoginRequested>(_onSmsLogin);
    on<AuthLogoutRequested>(_onLogout);
    on<AuthTokenRestored>(_onTokenRestored);
  }

  Future<void> _onLogin(AuthLoginRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    try {
      final result = await _repo.login(
        phone: event.phone,
        email: event.email,
        password: event.password,
      );

      if (result is Map && result['code'] == 0 && result['data'] != null) {
        final token = result['data']['access_token'] as String;
        final user = result['data']['user'] != null
            ? UserModel.fromJson(result['data']['user'])
            : UserModel(id: '', nickname: '用户', role: 'user');
        ApiClient().accessToken = token;
        SharedPreferences.getInstance().then((p) => p.setString('user_id', user.id));
        emit(AuthAuthenticated(user: user, accessToken: token));
      } else if (result is Map) {
        emit(AuthError(result['message'] as String? ?? '登录失败'));
      } else {
        emit(AuthError('登录失败: 服务器响应异常'));
      }
    } catch (e) {
      emit(AuthError('登录失败: ${e.toString()}'));
    }
  }

  Future<void> _onRegister(AuthRegisterRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    try {
      final result = await _repo.register(
        nickname: event.nickname,
        password: event.password,
        phone: event.phone,
        email: event.email,
      );

      if (result['code'] == 0) {
        emit(AuthUnauthenticated());
      } else {
        emit(AuthError(result['message'] as String? ?? '注册失败'));
      }
    } catch (e) {
      emit(AuthError('注册失败: ${e.toString()}'));
    }
  }

  Future<void> _onSmsLogin(AuthSmsLoginRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    try {
      final result = await _repo.smsLogin(event.phone, event.code, event.password, event.nickname);

      if (result['code'] == 0 && result['data'] != null) {
        final token = result['data']['access_token'] as String;
        final user = result['data']['user'] != null
            ? UserModel.fromJson(result['data']['user'])
            : UserModel(id: '', nickname: event.nickname, role: 'user');
        ApiClient().accessToken = token;
        SharedPreferences.getInstance().then((p) => p.setString('user_id', user.id));
        emit(AuthAuthenticated(user: user, accessToken: token));
      } else {
        emit(AuthError(result['message'] as String? ?? '注册失败'));
      }
    } catch (e) {
      emit(AuthError('注册失败: ${e.toString()}'));
    }
  }

  Future<void> _onLogout(AuthLogoutRequested event, Emitter<AuthState> emit) async {
    ApiClient().accessToken = null;
    emit(AuthUnauthenticated());
  }

  Future<void> _onTokenRestored(AuthTokenRestored event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    try {
      final result = await _repo.getCurrentUser();
      if (result['code'] == 0 && result['data'] != null) {
        final user = UserModel.fromJson(result['data']);
                SharedPreferences.getInstance().then((p) => p.setString('user_id', user.id));
        emit(AuthAuthenticated(user: user, accessToken: event.token));
      } else {
        emit(AuthUnauthenticated());
      }
    } catch (e) {
      emit(AuthUnauthenticated());
    }
  }
}

sealed class AuthEvent {}
class AuthLoginRequested extends AuthEvent {
  final String? phone;
  final String? email;
  final String password;
  AuthLoginRequested({this.phone, this.email, required this.password});
}
class AuthRegisterRequested extends AuthEvent {
  final String nickname;
  final String password;
  final String? phone;
  final String? email;
  AuthRegisterRequested({required this.nickname, required this.password, this.phone, this.email});
}
class AuthSmsLoginRequested extends AuthEvent {
  final String phone;
  final String code;
  final String nickname;
  final String password;
  AuthSmsLoginRequested({required this.phone, required this.code, required this.nickname, required this.password});
}
class AuthLogoutRequested extends AuthEvent {}
class AuthTokenRestored extends AuthEvent {
  final String token;
  AuthTokenRestored(this.token);
}
