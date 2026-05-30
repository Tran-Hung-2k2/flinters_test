trước tiên hãy đọc file readme.md mô tả yêu cầu bài toán này và tóm tắt cho tôi

tôi đang phân vân giữa golang và python cho bài toán này. Cái nào sẽ tối ưu hơn cho các yêu cầu của task này.

Giờ dựa trên yêu cầu trong file readme hãy viết code golang để thỏa mãn các yêu cầu đề bài. Đặc biệt là phải tối ưu về mặt hiệu suất nhanh và tối ưu bộ nhớ. Chương trình phải chạy được bằng CLI

chuyển hết các string đang hard code trong file main thành biến có thể nhận từ cli flags

Tối ưu thêm tốc độ parser CSV để đẩy hiệu năng lên cao hơn nữa. Có thể dùng goroutines để tăng tốc độ xử lý không. Nếu có thì viết Benchmark so sánh cách mới và cách cũ về các tiêu chí ví dụ như Peak memory usage, process time,...

Phần report đang dùng hàm sort của thư viện go. Giải thích hàm này dùng thuật toán sort nào. Và với bài toán này có cách nào tối ưu hơn không

Ok hãy thực hiện implement heap thay cho sort để tối ưu cho trường hợp nhiều campaign mà chỉ lấy top k rất nhỏ

Hãy thử kiến trúc sau: Một goroutine đọc CSV Scan byte-by-byte để tránh alloc không cần thiết, parse row (có thể Parse cents trực tiếp để tránh floating point), push vào channel. Sau đó Worker Pool (n goroutines) nhận row từ channel, aggregate local map để tránh race condition thì thao tác trên map global. Cuối cùng từ các local map hãy tổng hợp và đưa ra kết quả cuối cùng. Và hãy viết benchmark so sánh các tiêu chí ví dụ như Peak memory usage, process time, cpu usage, alloc mem,... trên toàn bộ tập dữ liệu cho 3 phương pháp

Hãy viết benchmark chi tiết hơn với nhiều thông số hơn

Hãy viết full test cho ứng dụng để kiểm tra output của chương trình có đáp ứng các tiêu chí của challenge không

hãy bổ sung test nhiều hơn coverage các case exception như không có file, malform row,... 

Từ các thông tin tôi đã cung cấp trong file readme.md hiện tại và source code hãy viết lại doc theo như yêu cầu của challenge vào file readmme.md
