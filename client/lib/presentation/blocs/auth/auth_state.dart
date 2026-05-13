import 'package:equatable/equatable.dart';
import 'package:nexusacg/core/models/models.dart';

sealed class AuthState extends Equatable {
  @override
  List<Object?> get props => [];
}

class AuthInitial extends AuthState {}
class AuthLoading extends AuthState {}
class AuthAuthenticated extends AuthState {
  final UserModel user;
  final String accessToken;
  AuthAuthenticated({required this.user, required this.accessToken});
  @override
  List<Object?> get props => [user, accessToken];
}
class AuthUnauthenticated extends AuthState {}
class AuthError extends AuthState {
  final String message;
  AuthError(this.message);
  @override
  List<Object?> get props => [message];
}
