type AudioProps = {
  src: string
  onError?: () => void
}
const AudioPlayer = ({ src, onError }: AudioProps) => {
  return (
    <audio 
      controls 
      onError={onError}
      style={{ width: '100%' }}
    >
      <source src={src} type="audio/mpeg" />
      Your browser does not support the audio element.
    </audio>
  )
}

export default AudioPlayer
