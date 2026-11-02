export default function Toast({ message }: { message: string | null }) {
  if (!message) return null;
  return (
    <div className="fixed bottom-6 right-6 bg-gray-900 text-white text-xs font-bold px-4 py-3 rounded-xl shadow-lg z-50">
      {message}
    </div>
  );
}
