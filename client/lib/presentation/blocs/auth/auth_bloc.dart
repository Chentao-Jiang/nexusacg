import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:nexusacg/core/network/api_client.dart';
import 'package:nexusacg/core/models/models.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/presentation/blocs/auth/auth_state.dart';

class AuthBloc extends Bloc<AuthEvent, AuthState> {
  final AuthRepository _repo = AuthRepository();

  AuthBloc() : super(AuthInitial()) {
    on<AuthLoginRequested>(_onLogin);
    on<AuthRegisterRequested>(_onRegister);
    on<AuthLogoutRequested>(_onLogout);
  }

  Future<void> _onLogin(AuthLoginRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    try {
      final result = await _repo.login(
        phone: event.phone,
        email: event.email,
        password: event.password,
      );

      if (result['code'] == 0 && result['data'] != null) {
        final token = result['data']['access_token'] as String;
        ApiClient().accessToken = token;
        emit(AuthUnauthenticated()); // Navigate to login success -> main
      } else {
        emit(AuthError(result['message'] as String? ?? '登录失败'));
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

  Future<void> _onLogout(AuthLogoutRequested event, Emitter<AuthState> emit) async {
    ApiClient().accessToken = null;
    emit(AuthUnauthenticated());
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
class AuthLogoutRequested extends AuthEvent {}
