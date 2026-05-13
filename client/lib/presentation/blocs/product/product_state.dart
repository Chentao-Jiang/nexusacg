import 'package:equatable/equatable.dart';
import 'package:nexusacg/core/models/models.dart';

sealed class ProductState extends Equatable {
  @override
  List<Object?> get props => [];
}

class ProductInitial extends ProductState {}
class ProductLoading extends ProductState {}
class ProductLoaded extends ProductState {
  final List<ProductModel> products;
  final bool hasMore;
  final String? zone;
  ProductLoaded({this.products = const [], this.hasMore = true, this.zone});
  @override
  List<Object?> get props => [products, hasMore, zone];
}
class ProductError extends ProductState {
  final String message;
  ProductError(this.message);
  @override
  List<Object?> get props => [message];
}
