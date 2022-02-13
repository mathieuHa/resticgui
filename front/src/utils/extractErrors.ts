const extractErrors = ({ inner }) => {
  return inner.reduce((acc, err) => {
    return { ...acc, [err.path]: err.message };
  }, {});
};

export default extractErrors;
