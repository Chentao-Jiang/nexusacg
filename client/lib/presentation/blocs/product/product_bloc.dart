import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:nexusacg/core/repositories/repositories.dart';
import 'package:nexusacg/presentation/blocs/product/product_state.dart';

class ProductBloc extends Bloc<ProductEvent, ProductState> {
  final ProductRepository _repo = ProductRepository();

  ProductBloc() : super(ProductInitial()) {
    on<ProductLoadRequested>(_onLoad);
    on<ProductLoadMoreRequested>(_onLoadMore);
    on<ProductSearchRequested>(_onSearch);
  }

  Future<void> _onLoad(ProductLoadRequested event, Emitter<ProductState> emit) async {
    emit(ProductLoading());
    try {
      final products = await _repo.getProducts(zone: event.zone, page: 1);
      emit(ProductLoaded(products: products, hasMore: products.length >= 20, zone: event.zone));
    } catch (e) {
      emit(ProductError('加载失败: ${e.toString()}'));
    }
  }

  Future<void> _onLoadMore(ProductLoadMoreRequested event, Emitter<ProductState> emit) async {
    if (state is! ProductLoaded) return;
    final currentState = state as ProductLoaded;
    if (!currentState.hasMore) return;

    try {
      final page = (currentState.products.length ~/ 20) + 1;
      final more = await _repo.getProducts(zone: currentState.zone, page: page);
      emit(ProductLoaded(
        products: [...currentState.products, ...more],
        hasMore: more.length >= 20,
        zone: currentState.zone,
      ));
    } catch (e) {
      // Keep current state on error
    }
  }

  Future<void> _onSearch(ProductSearchRequested event, Emitter<ProductState> emit) async {
    emit(ProductLoading());
    try {
      final products = await _repo.getProducts(keyword: event.keyword, page: 1);
      emit(ProductLoaded(products: products, hasMore: products.length >= 20));
    } catch (e) {
      emit(ProductError('搜索失败: ${e.toString()}'));
    }
  }
}

sealed class ProductEvent {}
class ProductLoadRequested extends ProductEvent {
  final String? zone;
  ProductLoadRequested({this.zone});
}
class ProductLoadMoreRequested extends ProductEvent {}
class ProductSearchRequested extends ProductEvent {
  final String keyword;
  ProductSearchRequested(this.keyword);
}
